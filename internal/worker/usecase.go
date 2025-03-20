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
		if err := u.runTask(ctx, m); err != nil {
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

func (u *UseCase) runTask(ctx context.Context, message *domain.QueueMessage) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	connector, err := u.connectorRepository.GetByID(ctx, message.ConnectorID)
	if err != nil {
		return err
	}

	// TODO: Implement Mappers
	connector.Mapper = domain.Mapper{
		ExternalID:  "uuid",
		Username:    "sAMAccountName",
		FullName:    "cn",
		PhoneNumber: "mobile",
		Email:       "mail",
		CreatedAt:   "whenCreated",
		UpdatedAt:   "whenChanged",
		Custom: map[string]string{
			"active": "accountExpires",
		},
	}

	count := 0
	switch message.TaskType {
	case domain.TaskTypeIncrementalSync:
		filter := fmt.Sprintf(
			"(%s>=%s)",
			connector.Mapper.UpdatedAt,
			connector.LastSync.Format("20060102150405.0Z"),
		)
		count, err = u.ldapSync(ctx, connector, filter)
	case domain.TaskTypeFullSync:
		count, err = u.ldapSync(ctx, connector)
	case domain.TaskTypeTerminate:
		u.terminateTask(message.ConnectorID)
		return nil
	}

	if err != nil {
		return err
	}

	connector.LastSync = time.Now()
	connector.UpdatedAt = connector.LastSync

	err = u.connectorRepository.UpdateSyncInfo(ctx, connector)
	if err != nil {
		return err
	}

	u.logger.Info("sync successfully", zap.Int("count", count))
	return nil
}

func (u *UseCase) ldapSync(ctx context.Context, connector *domain.Connector, filters ...string) (int, error) {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	connector.Data.Parsed = &domain.LdapConnector{}
	err := json.Unmarshal(connector.Data.Raw, connector.Data.Parsed)
	if err != nil {
		u.logger.Error("json.Unmarshal", zap.Error(err), zap.Any("raw", connector.Data.Raw))
		return 0, err
	}

	parsedConfig, ok := connector.Data.Parsed.(*domain.LdapConnector)
	if !ok {
		u.logger.Error(
			"connector.Data.Parsed.(*domain.LdapConnector)",
			zap.Bool("ok", ok),
			zap.Any("parsed", connector.Data.Parsed))
		return 0, errors.New("can not ")
	}

	conn, err := u.ldapClient.NewConnection(parsedConfig.URL, parsedConfig.ConnectTimeout)
	if err != nil {
		u.logger.Error("u.ldapClient.NewConnection", zap.Error(err))
		return 0, err
	}
	defer conn.Close()

	pwd, err := u.cipher.DecryptFromStdBase64(parsedConfig.SystemAccountPassword)
	if err != nil {
		u.logger.Error(
			"u.cipher.DecryptFromStdBase64",
			zap.Uint64("connector_id", connector.ConnectorID),
			zap.Error(err))
		return 0, err
	}

	err = conn.Bind(parsedConfig.SystemAccountDN, pwd)
	if err != nil {
		u.logger.Error("conn.Bind", zap.Error(err))
		return 0, err
	}

	count := 0
	pagingControl := ldap.NewControlPaging(parsedConfig.SyncSettings.BatchSize)
	for {
		resp, err := conn.Search(parsedConfig.BaseDN, filter,
			parsedConfig.ReadTimeout*time.Second,
			ldapclient.WithScope(ldap.ScopeSingleLevel),
			ldapclient.WithPagination(pagingControl))
		if err != nil {
			u.logger.Error("conn.Search", zap.Error(err))
			return 0, err
		}

		users := make([]*domain.User, len(resp.Entries))
		for i, entry := range resp.Entries {
			user := toUser(entry, connector.Mapper)
			user.SourceID = &connector.ConnectorID
			users[i] = user
		}

		if len(users) == 0 {
			break
		}

		_, err = u.userRepository.BulkInsertAndUpdate(ctx, users)
		if err != nil {
			u.logger.Error("u.userRepository.BulkInsertAndUpdate", zap.Error(err))
			return 0, err
		}

		count += len(resp.Entries)
		updatedControl := ldap.FindControl(resp.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			continue
		}
		break
	}

	return count, nil
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

func toUser(entry *ldap.Entry, mapper domain.Mapper) *domain.User {
	createdAt, _ := time.Parse("20060102150405Z", entry.GetAttributeValue(mapper.CreatedAt))
	updatedAt, _ := time.Parse("20060102150405Z", entry.GetAttributeValue(mapper.UpdatedAt))

	return &domain.User{
		FullName:    entry.GetAttributeValue(mapper.FullName),
		Username:    entry.GetAttributeValue(mapper.Username),
		PhoneNumber: entry.GetAttributeValue(mapper.PhoneNumber),
		Email:       entry.GetAttributeValue(mapper.Email),
		Active:      entry.GetAttributeValue(mapper.Custom["active"]) == "9223372036854775807",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
