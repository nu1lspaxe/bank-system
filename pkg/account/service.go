package account

import (
	"bank_system/postgres/sqlc"
	"bank_system/utils"
	"context"
)

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID int64) (*sqlc.BKAccount, error) {
	account, err := s.repo.CreateAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *AccountService) GetAccountByIDNumber(ctx context.Context, idNumber string) (*sqlc.GetAccountByIDNumberRow, error) {
	account, err := s.repo.GetAccountByIDNumber(ctx, idNumber)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *AccountService) GetAccountBalance(ctx context.Context, idNumber string) (float64, error) {
	account, err := s.repo.GetAccountByIDNumber(ctx, idNumber)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

func (s *AccountService) GetAccountTransactions(ctx context.Context, idNumber string) ([]sqlc.GetAccountTransactionsByIDNumberRow, error) {
	exists, err := s.repo.CheckAccountIDNumberExists(ctx, idNumber)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.NewBankSystemError(utils.ErrAccountNotFound)
	}

	transactions, err := s.repo.GetAccountTransactionsByIDNumber(ctx, idNumber)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *AccountService) GetAllAccounts(ctx context.Context) ([]sqlc.GetAllAccountsRow, error) {
	accounts, err := s.repo.GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *AccountService) Withdraw(ctx context.Context, idNumber string, amount float64, detail string) (int64, float64, error) {
	account, err := s.repo.GetAccountByIDNumber(ctx, idNumber)
	if err != nil {
		return 0, 0, err
	}
	if account.Balance < amount {
		return 0, 0, utils.NewBankSystemError(utils.ErrInsufficientBalance)
	}
	return s.repo.WithdrawFromAccount(ctx, account.ID, amount, detail)
}

func (s *AccountService) Deposit(ctx context.Context, accountID int64, amount float64, detail string) (int64, float64, error) {
	return s.repo.DepositToAccount(ctx, accountID, amount, detail)
}
