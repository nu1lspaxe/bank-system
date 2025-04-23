package user

import (
	"bank_system/postgres/sqlc"
	"bank_system/utils"
	"context"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*sqlc.BKUser, error) {
	exists, err := s.repo.CheckUserEmailExists(ctx, email)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, utils.NewBankSystemError(utils.ErrEmailExists)
	}

	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, username, email, hashPassword)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*sqlc.GetUserByIDRow, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserAccounts(ctx context.Context, id int64) (*[]sqlc.GetUserAccountsRow, error) {
	accounts, err := s.repo.GetUserAccounts(ctx, id)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) (*[]sqlc.GetAllUsersRow, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (s *UserService) CheckUserEmailExists(ctx context.Context, email string) (bool, error) {
	return s.repo.CheckUserEmailExists(ctx, email)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, username, email, password string) error {
	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	return s.repo.UpdateUser(ctx, id, username, email, hashPassword)
}
