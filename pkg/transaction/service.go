package transaction

import (
	"bank_system/postgres/sqlc"
	"context"
)

type TxService struct {
	repo TxRepository
}

func NewTxService(repo TxRepository) *TxService {
	return &TxService{
		repo: repo,
	}
}

func (s *TxService) GetTransactionByID(ctx context.Context, id int64) (*sqlc.BKTransaction, error) {
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
