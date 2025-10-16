package repository

import (
	"context"
	"duels-api/internal/model"
	"duels-api/pkg/repository"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type TransactionRepository struct {
	repository.Generic[model.TransactionType, string]
}

func NewTransactionRepository(
	generic repository.Generic[model.TransactionType, string],
) *TransactionRepository {
	return &TransactionRepository{Generic: generic}
}

func (r *TransactionRepository) WithTx(tx bun.Tx) *TransactionRepository {
	return &TransactionRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *TransactionRepository) BulkInsert(
	ctx context.Context,
	transactions []model.TransactionType,
) error {
	_, err := r.DB.NewInsert().
		Model(&transactions).
		On("CONFLICT (signature) DO NOTHING").
		Exec(ctx)
	return err
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	m *model.TransactionType,
) error {
	_, err := r.DB.NewInsert().
		Model(m).
		On("CONFLICT (signature) DO NOTHING").
		Exec(ctx)
	return err
}

func (r *TransactionRepository) BulkInsertWithSameTxType(
	ctx context.Context,
	txType uint8,
	signatures []string,
) error {
	_, err := r.DB.NewRaw(`
		INSERT INTO transactions (signature, tx_type)
		SELECT unnest(?::varchar(88)[]), ?
		ON CONFLICT DO NOTHING
	`, pgdialect.Array(signatures), txType).Exec(ctx)
	return err
}

func (r *TransactionRepository) GetTransactionsBySignatures(
	ctx context.Context,
	signatures []string,
) ([]model.TransactionType, error) {
	items := make([]model.TransactionType, 0, len(signatures))
	err := r.DB.NewSelect().
		Model(&items).
		Where("signature = ANY(?)", pgdialect.Array(signatures)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}
