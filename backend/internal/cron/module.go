package cron

import (
	rcron "github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("cron",
		fx.Provide(rcron.New),
		fx.Provide(NewSomeCron),
		fx.Provide(NewNotificationCron),
		fx.Invoke(
			func(lc fx.Lifecycle, cron *SomeCron) {
				lc.Append(fx.Hook{
					OnStart: cron.start,
					OnStop:  cron.stop,
				})
			},
			func(lc fx.Lifecycle, cron *NotificationCron) {
				lc.Append(fx.Hook{
					OnStart: cron.start,
					OnStop:  cron.stop,
				})
			},
		),
	)
}
