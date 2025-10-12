package repository

import (
	"duels-api/internal/model"
	"duels-api/pkg/repository"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("db",
		fx.Provide(
			CreateDBConnection,
		),
		fx.Provide(
			repository.NewTransactionManager,
		),
		fx.Provide(
			fx.Annotate(
				repository.NewDBWrapper,
				fx.As(
					new(repository.DB),
				)),
		),

		fx.Provide(
			repository.NewGenericRepository[model.User, uuid.UUID],
			NewUserRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Duel, uuid.UUID],
			NewDuelRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Player, uuid.UUID],
			NewPlayerRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.TransactionType, string],
			NewTransactionRepository,
		),
		fx.Provide(
			NewFileRepository,
		),
	)
}
