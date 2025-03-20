package pgrepo

import (
	"context"
	"github.com/google/uuid"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/utils"
)

type UserRepository struct {
	*db.PostgresClient
}

func NewUserRepository(pc *db.PostgresClient) *UserRepository {
	return &UserRepository{pc}
}

func (r *UserRepository) BulkUpsert(ctx context.Context, users []*domain.User) (int, error) {
	if len(users) == 0 {
		return 0, utils.ErrNoUserProvided
	}

	insertQuery := r.QueryBuilder.Insert(domain.TableUser).Columns(domain.AllUserSyncCols...)
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

	upsertQuery, args, err := insertQuery.Suffix(
		"ON CONFLICT (username) DO UPDATE " +
			"SET full_name = EXCLUDED.full_name, " +
			"phone_number = EXCLUDED.phone_number, " +
			"email = EXCLUDED.email, " +
			"data = EXCLUDED.data, " +
			"created_at = EXCLUDED.created_at, " +
			"updated_at = EXCLUDED.updated_at ",
	).ToSql()

	if err != nil {
		return 0, err
	}

	_, err = r.Pool.Exec(ctx, upsertQuery, args...)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
