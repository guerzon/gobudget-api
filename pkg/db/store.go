package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateUserParams, fn func(createdUser UserParams) error) (User, error)
	UpdateUserTx(ctx context.Context, arg UpdateUserParams, fn func(createdUser UserParams) error) (User, error)
	DeleteUserTx(ctx context.Context, userArg UserParams, budgetIds []uuid.UUID, afterDeleteFn func(deleteUser UserParams) error) error
	DeleteBudgetTx(ctx context.Context, budgetId uuid.UUID) error
	DeleteCategoryGroupTx(ctx context.Context, categoryGroupId uuid.UUID) error
}

type SQLStore struct {
	*Queries
	connPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
