package workeruc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/ldapclient"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

type UseCase struct {
	lock                *sync.Mutex
	runningTask         map[uint64]*domain.Task
	ldapClient          *ldapclient.LDAPClient
	connectorRepository *pgrepo.ConnectorRepository
	logger              *logger.ZapLogger
}

func NewUseCase(ldapClient *ldapclient.LDAPClient, connectorRepository *pgrepo.ConnectorRepository, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		lock:                new(sync.Mutex),
		runningTask:         make(map[uint64]*domain.Task),
		ldapClient:          ldapClient,
		connectorRepository: connectorRepository,
		logger:              zl,
	}
}

func (u *UseCase) RunTask(ctx context.Context, message domain.QueueMessage) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	u.lock.Lock()
	task, exist := u.runningTask[message.ConnectorID]

	if exist && isPrior(task.Type, message.TaskType) {
		u.lock.Unlock()
		u.logger.Error("a task with higher priority is currently running", zap.Any("task", task))
		return fmt.Errorf("higher priority task is running")
	}

	u.logger.Info("overriding the current task", zap.Any("task", task))
	u.terminateTask(message.ConnectorID)
	u.runningTask[message.ConnectorID] = &domain.Task{
		Type:      message.TaskType,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	defer delete(u.runningTask, message.ConnectorID)
	u.lock.Unlock()

	done := make(chan error)
	go func(m domain.QueueMessage) {
		defer close(done)
		if err := u.sync(ctx, m); err != nil {
			done <- err
		}
	}(message)

	select {
	case <-c.Done():
		return c.Err()
	case err := <-done: // also unblocked when done is closed
		return err
	}
}

func (u *UseCase) sync(ctx context.Context, message domain.QueueMessage) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	switch message.TaskType {
	case domain.TaskTypeIncrementalSync:
		return u.incrementalSync(ctx, message.ConnectorID)
	case domain.TaskTypeFullSync:
		return u.fullSync()
	case domain.TaskTypeTerminate:
		u.terminateTask(message.ConnectorID)
	}

	return nil
}

func (u *UseCase) incrementalSync(ctx context.Context, connectorID uint64) error {
	connector, err := u.connectorRepository.GetByID(ctx, connectorID)
	if err != nil {
		return err
	}

	conn, err := u.ldapClient.NewConnection("", 1*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = json.Unmarshal(connector.Data.Raw, &connector.Data.Parsed)
	if err != nil {
		return err
	}

	parsedConfig, ok := connector.Data.Parsed.(domain.LdapConnector)
	if !ok {
		return fmt.Errorf("invalid connector")
	}

	pagingControl := ldap.NewControlPaging(parsedConfig.SyncSettings.BatchSize)
	for {
		resp, err := conn.Search(parsedConfig.BaseDN,
			fmt.Sprintf("modifiedTimestamp > %s", connector.LastSync), // TODO: use mapping config to configure the real modifiedTimestamp attribute name
			parsedConfig.ReadTimeout,
			ldapclient.WithScope(ldap.ScopeSingleLevel),
			ldapclient.WithPagination(pagingControl))
		if err != nil {
			return err
		}

		updatedControl := ldap.FindControl(resp.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			continue
		}

		break
	}

	return nil
}

func (u *UseCase) fullSync() error {
	return nil
}

func (u *UseCase) terminateTask(connectorID uint64) {
	u.lock.Lock()
	defer u.lock.Unlock()

	job, exists := u.runningTask[connectorID]
	if exists && job.Cancel != nil {
		job.Cancel()
		delete(u.runningTask, connectorID)
	}
}

// isPrior return true if t2 has higher priority than t1
func isPrior(t1, t2 domain.TaskType) bool {
	switch t1 {
	case domain.TaskTypeIncrementalSync:
		return t2 == domain.TaskTypeFullSync || t2 == domain.TaskTypeTerminate
	case domain.TaskTypeFullSync:
		return t2 == domain.TaskTypeTerminate
	case domain.TaskTypeTerminate:
		return false
	default:
		return true
	}
}
