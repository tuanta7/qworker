package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/utils"
)

type ConnectorRepository struct {
	*db.PostgresClient
}

func NewConnectorRepository(pc *db.PostgresClient) *ConnectorRepository {
	return &ConnectorRepository{pc}
}

func (r *ConnectorRepository) List(ctx context.Context, page, pageSize uint64) ([]*domain.Connector, error) {
	query, args, err := r.PostgresClient.QueryBuilder.
		Select(domain.AllConnectorCols...).
		From(domain.TableConnector).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		ToSql()
	if err != nil {
		return nil, err
	}

	return r.list(ctx, query, args)
}

func (r *ConnectorRepository) ListByEnabled(ctx context.Context, enabled bool) ([]*domain.Connector, error) {
	query, args, err := r.PostgresClient.QueryBuilder.
		Select(domain.AllConnectorCols...).
		From(domain.TableConnector).
		Where(squirrel.Eq{domain.ColEnabled: enabled}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return r.list(ctx, query, args)
}

func (r *ConnectorRepository) GetByID(ctx context.Context, id uint64) (*domain.Connector, error) {
	query, args, err := r.PostgresClient.QueryBuilder.
		Select(domain.AllConnectorCols...).
		From(domain.TableConnector).
		Where(squirrel.Eq{domain.ColConnectorID: id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	c := &domain.Connector{}
	err = r.Pool.QueryRow(ctx, query, args...).Scan(
		&c.ConnectorID,
		&c.ConnectorType,
		&c.DisplayName,
		&c.Enabled,
		&c.LastSync,
		&c.Data,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrConnectorNotFound
		}
		return nil, err
	}

	// Active Directory
	c.Mapper = domain.Mapper{
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

	return c, nil
}

func (r *ConnectorRepository) UpdateSyncInfo(ctx context.Context, c *domain.Connector) error {
	query, args, err := r.PostgresClient.QueryBuilder.
		Update(domain.TableConnector).
		Set(domain.ColLastSync, c.LastSync).
		Set(domain.ColUpdatedAt, c.UpdatedAt).
		Where(squirrel.Eq{domain.ColConnectorID: c.ConnectorID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.Pool.Exec(ctx, query, args...)
	return err
}

func (r *ConnectorRepository) list(ctx context.Context, query string, args []interface{}) ([]*domain.Connector, error) {
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	connectors := make([]*domain.Connector, 0)
	for rows.Next() {
		var c domain.Connector
		err = rows.Scan(
			&c.ConnectorID,
			&c.ConnectorType,
			&c.DisplayName,
			&c.Enabled,
			&c.LastSync,
			&c.Data,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		connectors = append(connectors, &c)
	}

	return connectors, nil
}
