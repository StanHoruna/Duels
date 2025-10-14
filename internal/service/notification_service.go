package service

import (
	"context"
	"duels-api/internal/model"
	"duels-api/internal/storage/cache"
	"duels-api/internal/storage/repository"
	"duels-api/pkg/apperrors"
	repo "duels-api/pkg/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type NotificationService struct {
	DuelRepository         *repository.DuelRepository
	NotificationRepository *repository.NotificationRepository
	UserRepository         *repository.UserRepository
	PlayerRepository       *repository.PlayerRepository
	c                      *redis.Client
	Events                 *cache.EventPubSub
}

func NewNotificationService(
	duelRepository *repository.DuelRepository,
	notificationRepository *repository.NotificationRepository,
	userRepository *repository.UserRepository,
	playerRepository *repository.PlayerRepository,
	redisClient *redis.Client,
	notifications *cache.EventPubSub,
) (*NotificationService, error) {

	n := &NotificationService{
		DuelRepository:         duelRepository,
		NotificationRepository: notificationRepository,
		UserRepository:         userRepository,
		PlayerRepository:       playerRepository,
		c:                      redisClient,
		Events:                 notifications,
	}

	return n, nil
}

func (s *NotificationService) GetAllNotifications(
	ctx context.Context,
	userID uuid.UUID,
	options *repo.Options,
) ([]*model.Notification, error) {
	notifications, err := s.NotificationRepository.FindNotificationsByUserID(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to find user notifications by user id", err)
	}

	return notifications, nil
}

func (s *NotificationService) MarkAsRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) (*model.Notification, error) {
	notification, err := s.NotificationRepository.UpdateAsRead(ctx, userID, notificationID)
	if err != nil {
		return nil, apperrors.Internal("failed to update notification as read", err)
	}

	if notification == nil {
		return nil, apperrors.NotFound("user notification not found")
	}

	return notification, nil
}

func (s *NotificationService) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]*model.Notification, error) {
	notifications, err := s.NotificationRepository.UpdateAllAsRead(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to find user notifications by user id", err)
	}

	return notifications, nil
}

func (s *NotificationService) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	count, err := s.NotificationRepository.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, apperrors.Internal("failed to get unread notification count", err)
	}

	return count, nil
}

func (s *NotificationService) Publish(
	ctx context.Context,
	notification *model.Notification,
) error {
	err := s.NotificationRepository.Generic.Create(ctx, notification)
	if err != nil {
		return apperrors.Internal("failed to create notification", err)
	}

	err = s.c.Publish(ctx, s.Events.UserEventKey(notification.UserID), notification).Err()
	if err != nil {
		return apperrors.Internal("failed to publish notification", err)
	}

	return nil
}

func (s *NotificationService) CleanupOldNotifications(ctx context.Context, cutoff time.Time) error {
	err := s.NotificationRepository.DeleteOldNotifications(context.Background(), cutoff)
	if err != nil {
		return apperrors.Internal("failed to delete old notifications", err)
	}

	return nil
}
