package user

import (
	"bank_system/postgres/sqlc"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, email, password string) (sqlc.BKUser, error)
	GetUserByID(ctx context.Context, id int64) (sqlc.GetUserByIDRow, error)
	GetUserAccounts(ctx context.Context, id int64) ([]sqlc.GetUserAccountsRow, error)
	GetAllUsers(ctx context.Context) ([]sqlc.GetAllUsersRow, error)
	CheckUserEmailExists(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, id int64, username, email, password string) error
}

type userRepositoryImpl struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepositoryImpl{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, username, email, password string) (sqlc.BKUser, error) {
	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username: username,
		Email:    email,
		Password: password,
	})
}

func (r *userRepositoryImpl) GetUserByID(ctx context.Context, id int64) (sqlc.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *userRepositoryImpl) GetUserAccounts(ctx context.Context, id int64) ([]sqlc.GetUserAccountsRow, error) {
	return r.queries.GetUserAccounts(ctx, id)
}

func (r *userRepositoryImpl) GetAllUsers(ctx context.Context) ([]sqlc.GetAllUsersRow, error) {
	users, err := r.queries.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryImpl) CheckUserEmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.CheckUserEmailExists(ctx, email)
}

func (r *userRepositoryImpl) UpdateUser(ctx context.Context, id int64, username, email, password string) error {
	_, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       id,
		Username: username,
		Email:    email,
		Password: password,
	})
	return err
}
