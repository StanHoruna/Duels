package cron

import (
	"context"
	"duels-api/pkg/apperrors"
	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type SomeCron struct {
	Log  *zap.Logger
	Cron *rcron.Cron
}

const (
	RunningDailyAt11PM = "0 23 * * *"
)

func NewSomeCron(
	l *zap.Logger,
	cron *rcron.Cron,
	// add any additional service
) (*SomeCron, error) {
	someCron := &SomeCron{
		Log:  l,
		Cron: cron,
	}

	// usual time based cron job
	//_, err := coinCron.Cron.AddFunc(RunningDailyAt11PM, coinCron.doSomething)
	//if err != nil {
	//	return nil, err
	//}

	//coinCron.doSomething() // uncomment to start job after launch

	return someCron, nil
}

func (c *SomeCron) doSomething() {
	err := error(nil) // call any function
	if err != nil {
		LogErr(c.Log, err)
	} else {
		c.Log.Debug("some-job cron: successfully finished")
	}
}

func (c *SomeCron) start(_ context.Context) error {
	c.Log.Info("some-job cron started")
	c.Cron.Start()
	return nil
}

func (c *SomeCron) stop(_ context.Context) error {
	c.Log.Info("some-job cron stopped")
	c.Cron.Stop()
	return nil
}

func LogErr(l *zap.Logger, err error) {
	appErr, ok := apperrors.IsAppError(err)
	if !ok {
		l.Error("cron job failed", zap.Error(err))
		return
	}

	if appErr.BaseError != nil {
		l.Error(
			appErr.Message,
			zap.String("err", appErr.BaseError.Error()),
			zap.String("occurred", appErr.Path()),
		)
	} else {
		l.Error(
			appErr.Message,
			zap.String("occurred", appErr.Path()),
		)
	}
}
