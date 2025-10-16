package repository

import (
	"context"
	"database/sql"
	"duels-api/internal/model"
	"duels-api/pkg/repository"
	"errors"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type NotificationRepository struct {
	repository.Generic[model.Notification, uuid.UUID]
}

func NewNotificationRepository(
	genericRepository repository.Generic[model.Notification, uuid.UUID],
) *NotificationRepository {
	return &NotificationRepository{Generic: genericRepository}
}

func (r *NotificationRepository) WithTx(tx bun.Tx) *NotificationRepository {
	return &NotificationRepository{
		Generic: r.Generic.WithTx(tx),
	}
}

func (r *NotificationRepository) FindNotificationsByUserID(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]*model.Notification, error) {
	var notifs []*model.Notification

	subQ := r.DB.NewSelect().Model(&notifs)

	subQ = options.ApplyFilters(subQ)
	subQ = options.ApplyOrderBy(subQ)

	q := r.DB.NewSelect().
		ColumnExpr("sub.*").
		TableExpr("(?) AS sub", subQ).
		Where("sub.user_id = ?", userID)

	q = options.ApplyPagination(q)

	err := q.Scan(ctx, &notifs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return notifs, nil
}

func (r *NotificationRepository) UpdateAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]*model.Notification, error) {
	var notifs []*model.Notification

	err := r.DB.NewUpdate().
		Model(&notifs).
		Set("is_read = ?", true).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Returning("*").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return notifs, nil
}

func (r *NotificationRepository) UpdateAsRead(
	ctx context.Context,
	userID uuid.UUID,
	notifID uuid.UUID,
) (*model.Notification, error) {
	var notif model.Notification

	err := r.DB.NewUpdate().
		Model(&notif).
		Set("is_read = ?", true).
		Where("id = ?", notifID).
		Where("user_id = ?", userID).
		Returning("*").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &notif, nil
}

func (r *NotificationRepository) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	count, err := r.DB.NewSelect().
		Model((*model.Notification)(nil)).
		Where("user_id = ?", userID).
		Where("is_read = false").
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *NotificationRepository) DeleteOldNotifications(
	ctx context.Context,
	cutoff time.Time,
) error {
	_, err := r.DB.NewDelete().
		Model((*model.Notification)(nil)).
		Where("created_at < ?", cutoff).
		Exec(ctx)

	return err
}
