package workeruc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/cipherx"
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
	cipher              cipherx.Cipher
	connectorRepository *pgrepo.ConnectorRepository
	userRepository      *pgrepo.UserRepository
	logger              *logger.ZapLogger
}

func NewUseCase(
	ldapClient *ldapclient.LDAPClient,
	cipher cipherx.Cipher,
	connectorRepository *pgrepo.ConnectorRepository,
	userRepository *pgrepo.UserRepository,
	zl *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		lock:                &sync.Mutex{},
		runningTask:         make(map[uint64]*domain.Task),
		ldapClient:          ldapClient,
		cipher:              cipher,
		connectorRepository: connectorRepository,
		userRepository:      userRepository,
		logger:              zl,
	}
}

func (u *UseCase) RunTask(ctx context.Context, message *domain.QueueMessage) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	u.lock.Lock()
	task, exist := u.runningTask[message.ConnectorID]

	if exist {
		if isPrior(message.TaskType, task.Type) {
			u.lock.Unlock()
			u.logger.Error("a task with equal or higher priority is currently running", zap.Any("task", task))
			return fmt.Errorf("equal or higher priority task is running")
		}
		u.logger.Info("overriding the current task",
			zap.Any("currentTaskType", task.Type),
			zap.Any("newTaskType", message.TaskType),
		)
		u.terminateTask(message.ConnectorID)
	}

	u.runningTask[message.ConnectorID] = &domain.Task{
		Type:      message.TaskType,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	defer delete(u.runningTask, message.ConnectorID)
	u.lock.Unlock()

	done := make(chan error)
	go func(ctx context.Context, m *domain.QueueMessage) {
		defer close(done)
		if err := u.sync(ctx, m); err != nil {
			done <- err
		}
	}(c, message)

	select {
	case <-c.Done():
		return c.Err()
	case err := <-done: // also unblocked when done is closed
		return err
	}
}

func (u *UseCase) sync(ctx context.Context, message *domain.QueueMessage) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	connector, err := u.connectorRepository.GetByID(ctx, message.ConnectorID)
	if err != nil {
		return err
	}

	switch message.TaskType {
	case domain.TaskTypeIncrementalSync:
		err = u.incrementalSync(ctx, connector)
	case domain.TaskTypeFullSync:
		err = u.fullSync(ctx, connector)
	case domain.TaskTypeTerminate:
		u.terminateTask(message.ConnectorID)
		return nil
	}

	if err != nil {
		return err
	}

	connector.LastSync = time.Now()

	return nil
}

func (u *UseCase) incrementalSync(ctx context.Context, connector *domain.Connector) error {
	// TODO: detect sync method based on the connector type

	connector.Data.Parsed = &domain.LdapConnector{}
	err := json.Unmarshal(connector.Data.Raw, connector.Data.Parsed)
	if err != nil {
		u.logger.Error("unable to unmarshal LDAP connector data", zap.Error(err))
		return err
	}

	parsedConfig, ok := connector.Data.Parsed.(*domain.LdapConnector)
	if !ok {
		u.logger.Error("unable to unmarshal LDAP connector data", zap.Any("connector", connector))
		return errors.New("invalid connector")
	}

	conn, err := u.ldapClient.NewConnection(parsedConfig.URL, parsedConfig.ConnectTimeout)
	if err != nil {
		u.logger.Error("unable to create LDAP connection", zap.Error(err))
		return err
	}
	defer conn.Close()

	pwd, err := u.cipher.DecryptFromStdBase64(parsedConfig.SystemAccountPassword)
	if err != nil {
		u.logger.Error("unable to decrypt LDAP system account password", zap.Error(err))
		return err
	}

	err = conn.Bind(parsedConfig.SystemAccountDN, pwd)
	if err != nil {
		u.logger.Error("unable to bind LDAP connection", zap.Error(err))
		return err
	}

	// TODO:
	connector.Mapper = domain.Mapper{
		ExternalID:  "uuid",
		FullName:    "cn",
		PhoneNumber: "mobile",
		Email:       "mail",
		CreatedAt:   "whenCreated",
		UpdatedAt:   "whenChanged",
		Custom: map[string]string{
			"active": "accountExpires",
		},
	}

	filter := fmt.Sprintf("(%s>=%s)", connector.Mapper.UpdatedAt, connector.LastSync.Format("20060102150405.0Z"))
	pagingControl := ldap.NewControlPaging(parsedConfig.SyncSettings.BatchSize)

	for {
		resp, err := conn.Search(parsedConfig.BaseDN,
			filter,
			parsedConfig.ReadTimeout*time.Second,
			ldapclient.WithScope(ldap.ScopeSingleLevel),
			ldapclient.WithPagination(pagingControl))
		if err != nil {
			u.logger.Error("unable to search LDAP", zap.Error(err))
			return err
		}

		u.logger.Info(
			"entries",
			zap.Any("count", len(resp.Entries)),
			zap.Any("control", resp.Controls),
		)

		users := make([]*domain.User, len(resp.Entries))
		for i, e := range resp.Entries {
			user := toUser(e, parsedConfig.UsernameAttribute, connector.Mapper)
			user.SourceID = &connector.ConnectorID
			users[i] = user
		}

		_, err = u.userRepository.BulkInsertAndUpdate(ctx, users)
		if err != nil {
			u.logger.Error("unable to bulk insert user", zap.Error(err))
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

func (u *UseCase) fullSync(ctx context.Context, connector *domain.Connector) error {
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

// isPrior check if t2 has higher priority than t1
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

func toUser(entry *ldap.Entry, usernameAttribute string, mapper domain.Mapper) *domain.User {
	createdAt, _ := time.Parse("20060102150405Z", entry.GetAttributeValue(mapper.CreatedAt))
	updatedAt, _ := time.Parse("20060102150405Z", entry.GetAttributeValue(mapper.UpdatedAt))

	return &domain.User{
		FullName:    entry.GetAttributeValue(mapper.FullName),
		Username:    entry.GetAttributeValue(usernameAttribute),
		PhoneNumber: entry.GetAttributeValue(mapper.PhoneNumber),
		Email:       entry.GetAttributeValue(mapper.Email),
		Active:      entry.GetAttributeValue(mapper.Custom["active"]) == "9223372036854775807",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
