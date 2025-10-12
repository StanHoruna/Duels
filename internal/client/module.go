package client

import (
	"duels-api/internal/client/solana"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("Clients",
		fx.Provide(
			solana.NewClient,
		),
	)
}
