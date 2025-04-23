package transaction

import (
	"bank_system/postgres/sqlc"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TxRepository interface {
	CreateTransaction(ctx context.Context, accountID int64, amount float64, txType, detail string) (sqlc.BKTransaction, error)
	GetTransactionByID(ctx context.Context, id int64) (sqlc.BKTransaction, error)
}

type txRepistoryImpl struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewTxRepository(pool *pgxpool.Pool) TxRepository {
	return &txRepistoryImpl{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *txRepistoryImpl) CreateTransaction(ctx context.Context, accountID int64, amount float64, txType, detail string) (sqlc.BKTransaction, error) {
	tx, err := r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		AccountID: accountID,
		Amount:    amount,
		TxType:    txType,
		Detail:    detail,
	})
	if err != nil {
		return sqlc.BKTransaction{}, err
	}
	return tx, nil
}

func (r *txRepistoryImpl) GetTransactionByID(ctx context.Context, id int64) (sqlc.BKTransaction, error) {
	return r.queries.GetTransactionByID(ctx, id)
}
