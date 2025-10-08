package repository

import (
	"context"
	"duels-api/internal/model"
	"duels-api/pkg/repository"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	repository.Generic[model.User, uuid.UUID]
}

func (r *UserRepository) WithTx(tx bun.Tx) *UserRepository {
	return &UserRepository{Generic: r.Generic.WithTx(tx)}
}

func NewUserRepository(
	genericRepository repository.Generic[model.User, uuid.UUID],
) *UserRepository {
	return &UserRepository{
		Generic: genericRepository,
	}
}

func (r *UserRepository) GetByPublicAddress(
	ctx context.Context,
	address string,
) (*model.User, error) {
	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("public_address = ?", address).
		Scan(ctx)
	if err != nil {
		if repository.IsErrNoRows(err) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
