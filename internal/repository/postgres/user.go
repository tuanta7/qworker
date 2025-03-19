package pgrepo

import (
	"context"
	"github.com/google/uuid"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
)

type UserRepository struct {
	*db.PostgresClient
}

func NewUserRepository(pc *db.PostgresClient) *UserRepository {
	return &UserRepository{pc}
}

func (r *UserRepository) BulkInsertAndUpdate(ctx context.Context, users []*domain.User) (int, error) {
	insertQuery := r.QueryBuilder.Insert(domain.TableUser).Columns(domain.AllUserCols...)

	for _, user := range users {
		insertQuery = insertQuery.Values(
			uuid.NewString(),
			user.Username,
			user.FullName,
			user.PhoneNumber,
			user.Email,
			user.EmailVerified,
			user.Active,
			user.SourceID,
			user.Data,
			user.CreatedAt,
			user.UpdatedAt,
		)
	}

	query, args, err := insertQuery.ToSql()
	if err != nil {
		return 0, err
	}

	_, err = r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
