package app

import (
	"duels-api/config"
	"duels-api/internal/client"
	"duels-api/internal/cron"
	"duels-api/internal/handler/server"
	v1 "duels-api/internal/handler/v1"
	"duels-api/internal/service"
	"duels-api/internal/storage/cache"
	"duels-api/internal/storage/repository"
	auth "duels-api/pkg/jwt"
	"duels-api/pkg/logger"
	"go.uber.org/fx"
)

func Build() *fx.App {
	return fx.New(
		fx.Options(
			config.Module,
			logger.Module,
		),
		auth.Module(),

		repository.Module(),
		cache.Module(),

		client.Module(),

		service.Module(),
		server.Module(),

		v1.Module(),
		cron.Module(),
	)
}
