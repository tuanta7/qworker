package workeruc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/pkg/cipherx"
	"github.com/tuanta7/qworker/pkg/ldapclient"
	"github.com/tuanta7/qworker/pkg/logger"
	"github.com/tuanta7/qworker/pkg/utils"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type UseCase struct {
	asynqInspector      *asynq.Inspector
	ldapClient          *ldapclient.LDAPClient
	cipher              cipherx.Cipher
	connectorRepository *pgrepo.ConnectorRepository
	userRepository      *pgrepo.UserRepository
	logger              *logger.ZapLogger
}

func NewUseCase(
	asynqInspector *asynq.Inspector,
	ldapClient *ldapclient.LDAPClient,
	cipher cipherx.Cipher,
	connectorRepository *pgrepo.ConnectorRepository,
	userRepository *pgrepo.UserRepository,
	zl *logger.ZapLogger,
) *UseCase {
	return &UseCase{
		asynqInspector:      asynqInspector,
		ldapClient:          ldapClient,
		cipher:              cipher,
		connectorRepository: connectorRepository,
		userRepository:      userRepository,
		logger:              zl,
	}
}

func (u *UseCase) IsConnectorRunning(connectorID uint64) (*asynq.TaskInfo, error) {
	for queue := range domain.QueuePriority {
		taskID := fmt.Sprintf("asynq{%s}:t:%d", queue, connectorID)

		tasks, err := u.asynqInspector.ListActiveTasks(queue) // get max 30 tasks by default
		if err != nil {
			u.logger.Error("u.asynqInspector.ListActiveTasks", zap.Error(err), zap.String("queue", queue))
			if errors.Is(err, asynq.ErrQueueNotFound) {
				continue
			}
			return nil, err
		}

		for _, task := range tasks {
			if task.ID == taskID {
				return task, nil
			}
		}
	}

	return nil, nil
}

func (u *UseCase) TerminateTask(queue string, connectorID uint64) error {
	return nil
}

func (u *UseCase) CleanTask(queue string, connectorID uint64) error {
	return u.asynqInspector.DeleteTask(queue, strconv.FormatUint(connectorID, 10))
}

func (u *UseCase) RunTask(ctx context.Context, message *domain.QueueMessage) error {
	connector, err := u.connectorRepository.GetByID(ctx, message.ConnectorID)
	if err != nil {
		return err
	}

	// TODO: Implement attribute mappers
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
	switch connector.ConnectorType {
	case domain.ConnectorTypeLDAP:
		count, err = u.runLdapSyncTask(ctx, message, connector)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported connector type")
	}

	connector.LastSync = time.Now()
	connector.UpdatedAt = connector.LastSync

	err = u.connectorRepository.UpdateSyncInfo(ctx, connector)
	if err != nil {
		return err
	}

	u.logger.Info("user synced successfully", zap.Int("count", count))
	return nil
}

func (u *UseCase) runLdapSyncTask(ctx context.Context, msg *domain.QueueMessage, c *domain.Connector) (count int, err error) {
	switch msg.TaskType {
	case domain.TaskTypeIncrementalSync:
		filter := fmt.Sprintf(
			"(%s>=%s)",
			c.Mapper.UpdatedAt,
			c.LastSync.Format("20060102150405.0Z"),
		)
		count, err = u.ldapSync(ctx, c, filter)
	case domain.TaskTypeFullSync:
		count, err = u.ldapSync(ctx, c)
	case domain.TaskTypeTerminate:
		// u.TerminateTask(msg.ConnectorID)
		return 0, nil
	}
	return count, nil
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
		return 0, errors.New("cannot assert data to type *domain.LdapConnector")
	}

	conn, err := u.ldapClient.NewConnection(parsedConfig.URL, parsedConfig.ConnectTimeout*time.Millisecond)
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

		_, err = u.userRepository.BulkUpsert(ctx, users)
		if err != nil {
			if errors.Is(err, utils.ErrNoUserProvided) {
				u.logger.Info(err.Error())
				break
			}
			u.logger.Error("u.userRepository.BulkUpsert", zap.Error(err))
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
