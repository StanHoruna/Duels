package cron

import (
	"context"
	"duels-api/config"
	"duels-api/internal/service"
	"time"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type NotificationCron struct {
	Log                 *zap.Logger
	Cron                *rcron.Cron
	DuelService         *service.DuelService
	NotificationService *service.NotificationService
	NotificationTTL     uint32
}

const (
	RunningDailyAt12AM = "0 0 * * *"
)

func NewNotificationCron(
	c *config.Config,
	l *zap.Logger,
	cron *rcron.Cron,
	duelService *service.DuelService,
	notificationService *service.NotificationService,
) (*NotificationCron, error) {
	notificationCron := &NotificationCron{
		Log:                 l,
		Cron:                cron,
		DuelService:         duelService,
		NotificationService: notificationService,
		NotificationTTL:     c.App.NotificationTtl,
	}

	_, err := notificationCron.Cron.AddFunc(RunningDailyAt12AM, notificationCron.cleanupOldNotification)
	if err != nil {
		return nil, err
	}

	return notificationCron, nil
}

func (c *NotificationCron) cleanupOldNotification() {
	cutoff := time.Now().Add(-24 * time.Hour * time.Duration(c.NotificationTTL))
	err := c.NotificationService.CleanupOldNotifications(context.Background(), cutoff)
	if err != nil {
		LogErr(c.Log, err)
	} else {
		c.Log.Debug("notification cron: successfully cleaned up old notifications")
	}
}

func (c *NotificationCron) start(_ context.Context) error {
	c.Log.Info("notification cron started")
	c.Cron.Start()
	return nil
}

func (c *NotificationCron) stop(_ context.Context) error {
	c.Log.Info("notification cron stopped")
	c.Cron.Stop()
	return nil
}
