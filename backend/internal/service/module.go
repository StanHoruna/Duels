package service

import (
	"context"
	"duels-api/config"
	"duels-api/pkg/sigtracker"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("services",
		fx.Provide(
			NewUserService,
			NewJWTService,
			NewFileService,
			NewDuelService,
			NewWalletService,
			NewPriorityTracker,
			NewNotificationService,
		),
		fx.Provide(
			func(lc fx.Lifecycle, client *rpc.Client, cfg *config.Config) *sigtracker.TxTracker {
				tracker := sigtracker.NewTransactionTracker(client, cfg.App.SolanaWSNodeURL)

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						tracker.Start()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return tracker.Close()
					},
				})

				return tracker
			},
		),
	)
}
