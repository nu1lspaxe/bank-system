package account

import (
	"bank_system/postgres/sqlc"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, userID int64) (sqlc.BKAccount, error)
	CheckAccountIDNumberExists(ctx context.Context, idNumber string) (bool, error)
	GetAccountByIDNumber(ctx context.Context, idNumber string) (sqlc.GetAccountByIDNumberRow, error)
	GetAccountTransactionsByIDNumber(ctx context.Context, idNumber string) ([]sqlc.GetAccountTransactionsByIDNumberRow, error)
	GetAllAccounts(ctx context.Context) ([]sqlc.GetAllAccountsRow, error)
	WithdrawFromAccount(ctx context.Context, accountID int64, amount float64, detail string) (int64, float64, error)
	DepositToAccount(ctx context.Context, accountID int64, amount float64, detail string) (int64, float64, error)
}

type accountRepositoryImpl struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) AccountRepository {
	return &accountRepositoryImpl{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *accountRepositoryImpl) CreateAccount(ctx context.Context, userID int64) (sqlc.BKAccount, error) {
	tx, err := r.queries.CreateAccount(ctx, userID)
	if err != nil {
		return sqlc.BKAccount{}, err
	}
	return tx, nil
}

func (r *accountRepositoryImpl) CheckAccountIDNumberExists(ctx context.Context, idNumber string) (bool, error) {
	exists, err := r.queries.CheckAccountIDNumberExists(ctx, idNumber)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *accountRepositoryImpl) GetAccountByIDNumber(
	ctx context.Context, idNumber string,
) (sqlc.GetAccountByIDNumberRow, error) {
	return r.queries.GetAccountByIDNumber(ctx, idNumber)
}

func (r *accountRepositoryImpl) GetAccountTransactionsByIDNumber(
	ctx context.Context, idNumber string,
) ([]sqlc.GetAccountTransactionsByIDNumberRow, error) {
	return r.queries.GetAccountTransactionsByIDNumber(ctx, idNumber)
}

func (r *accountRepositoryImpl) GetAllAccounts(ctx context.Context) ([]sqlc.GetAllAccountsRow, error) {
	accounts, err := r.queries.GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepositoryImpl) WithdrawFromAccount(
	ctx context.Context, accountID int64, amount float64, detail string,
) (int64, float64, error) {
	txOptions := pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	}

	tx, err := r.pool.BeginTx(ctx, txOptions)

	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	result, err := r.queries.WithTx(tx).WithdrawFromAccount(ctx, sqlc.WithdrawFromAccountParams{
		AccountID: accountID,
		Amount:    amount,
		TxDetail:  detail,
	})

	if err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, err
	}

	return result.TransactionID, result.NewBalance, nil
}

func (r *accountRepositoryImpl) DepositToAccount(
	ctx context.Context, accountID int64, amount float64, detail string,
) (int64, float64, error) {
	txOptions := pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	}

	tx, err := r.pool.BeginTx(ctx, txOptions)

	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	result, err := r.queries.WithTx(tx).DepositToAccount(ctx, sqlc.DepositToAccountParams{
		AccountID: accountID,
		Amount:    amount,
		TxDetail:  detail,
	})

	if err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, err
	}

	return result.TransactionID, result.NewBalance, nil
}
