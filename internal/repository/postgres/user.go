package pgrepo

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
)

type UserRepository struct {
	db.PostgresClient
}

func NewUserRepository(pc db.PostgresClient) *UserRepository {
	return &UserRepository{pc}
}

func (r *UserRepository) BuildBulkUpsertQuery(users []*domain.User) *squirrel.InsertBuilder {
	if len(users) == 0 {
		return nil
	}

	insertQuery := r.QueryBuilder().Insert(domain.TableUser).Columns(domain.AllUserSyncCols...)
	for _, user := range users {
		insertQuery = insertQuery.Values(
			uuid.NewString(),
			user.Username,
			user.FullName,
			user.PhoneNumber,
			user.Email,
			user.SourceID,
			user.Data,
			user.CreatedAt,
			user.UpdatedAt,
		)
	}

	upsertQuery := insertQuery.Suffix(
		"ON CONFLICT (username) DO UPDATE " +
			"SET full_name = EXCLUDED.full_name, " +
			"phone_number = EXCLUDED.phone_number, " +
			"email = EXCLUDED.email, " +
			"data = EXCLUDED.data, " +
			"source_id = EXCLUDED.source_id, " +
			"created_at = EXCLUDED.created_at, " +
			"updated_at = EXCLUDED.updated_at ",
	)

	return &upsertQuery
}

func (r *UserRepository) ExecuteTransaction(ctx context.Context, queries []squirrel.Sqlizer) error {
	tx, err := r.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, query := range queries {
		sqlStr, args, err := query.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sqlStr, args...)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
