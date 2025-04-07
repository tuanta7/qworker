package workeruc

import (
	"context"
	"errors"
	"fmt"
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

func (u *UseCase) GetTask(id uint64, queue string) (*asynq.TaskInfo, error) {
	taskInfo, err := u.asynqInspector.GetTaskInfo(queue, strconv.FormatUint(id, 10))
	if err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) || errors.Is(err, asynq.ErrQueueNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return taskInfo, nil
}

func (u *UseCase) RunIncrementalSyncTask(ctx context.Context, message *domain.QueueMessage) error {
	c, err := u.connectorRepository.GetByID(ctx, message.ConnectorID)
	if err != nil {
		return err
	}

	if !c.Enabled {
		return errors.New("connector is disabled")
	}

	syncSettings, err := c.GetSyncSettings()
	if err != nil {
		return err
	}

	if !syncSettings.IncSync {
		return errors.New("incremental sync is disabled")
	}

	switch c.ConnectorType {
	case domain.ConnectorTypeLDAP:
		filter := fmt.Sprintf("(%s>=%s)", c.Mapper.UpdatedAt, utils.TimeToLDAPString(c.LastSync))
		err = u.ldapSync(ctx, c, filter)
	default:
		return errors.New("unsupported connector type")
	}
	if err != nil {
		return err
	}

	c.LastSync = time.Now()
	c.UpdatedAt = c.LastSync
	err = u.connectorRepository.UpdateSyncInfo(ctx, c)
	if err != nil {
		u.logger.Error("RunIncrementalSyncTask - u.connectorRepository.UpdateSyncInfo", zap.Error(err))
		return err
	}

	return nil
}

func (u *UseCase) RunFullSyncTask(ctx context.Context, message *domain.QueueMessage) error {
	c, err := u.connectorRepository.GetByID(ctx, message.ConnectorID)
	if err != nil {
		return err
	}

	if !c.Enabled {
		return errors.New("connector is disabled")
	}

	switch c.ConnectorType {
	case domain.ConnectorTypeLDAP:
		err = u.ldapSync(ctx, c)
	default:
		return errors.New("unsupported connector type")
	}
	if err != nil {
		return err
	}

	c.LastSync = time.Now()
	c.UpdatedAt = c.LastSync
	err = u.connectorRepository.UpdateSyncInfo(ctx, c)
	if err != nil {
		u.logger.Error("RunIncrementalSyncTask - u.connectorRepository.UpdateSyncInfo", zap.Error(err))
		return err
	}

	return nil
}
