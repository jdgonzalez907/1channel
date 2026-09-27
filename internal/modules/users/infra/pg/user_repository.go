package pg

import (
	"context"
	"errors"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

type postgresUserRepository struct {
	db *pgdb.DB
}

func NewUserRepository(db *pgdb.DB) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.db.Queries.FindUserByID(ctx, pgdb.UUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return domain.RehydrateUser(pgdb.FromUUID(row.ID), pgdb.FromTimestamp(row.CreatedAt)), nil
}

func (r *postgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	return r.db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        pgdb.UUID(user.ID()),
		CreatedAt: pgdb.Timestamp(user.CreatedAt()),
	})
}
