package service

import (
	"context"
	"duels-api/internal/model"
	"duels-api/pkg/apperrors"
	"fmt"
	"github.com/google/uuid"
)

func (s *DuelService) hasChargedDuelPriceFromUser(duelPlayersCount uint64, duelOldStatus, duelNewStatus uint8) bool {
	duelIsInProcess := duelOldStatus == model.DuelStatusInProcess
	isNewStatusRefund := duelNewStatus == model.DuelStatusRefund

	duelIsInReview := duelOldStatus == model.DuelStatusInReview
	isNewStatusAdminCancelled := duelNewStatus == model.DuelStatusAdminCancelled

	if ((duelIsInProcess && isNewStatusRefund) ||
		(duelIsInReview && isNewStatusAdminCancelled)) &&
		duelPlayersCount > 0 {

		return true
	}

	return false
}

func (s *DuelService) isAbleToJoinDuel(ctx context.Context, duel *model.Duel, userID uuid.UUID) error {
	duelIsInProgress := duel.Status == model.DuelStatusInProcess
	duelIsInReview := duel.Status == model.DuelStatusInReview

	if !(duelIsInProgress || duelIsInReview) {
		return apperrors.BadRequest("failed to join duel")
	}

	isPlayer, err := s.PlayerRepository.UserAlreadyParticipant(ctx, userID, duel.ID)
	if err != nil {
		return apperrors.Internal("failed to check if user is already participating in duel", err)
	}

	if isPlayer {
		return apperrors.BadRequest("user is already participating in this duel")
	}

	return nil
}

func (s *DuelService) validateResolveDuelByOwner(ownerID uuid.UUID, duel *model.Duel) error {
	if duel.OwnerID != ownerID {
		return apperrors.BadRequest("only the owner of the duel can resolve it")
	}

	if duel.Status != model.DuelStatusInProcess {
		return apperrors.BadRequest("resolve is not possible from current status")
	}

	return nil
}

func (s *DuelService) sendDuelShareImageReq(duel *model.Duel) error {
	resp, err := s.HTTPShareClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(duel).
		Post(s.ShareImageAPI + "duel")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("status: %d, body: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

func (s *DuelService) sendNotification(
	ctx context.Context,
	userID uuid.UUID,
	notificationType uint8,
	notificationPayload model.NotificationPayload,
) error {
	notification, err := model.NewNotification(
		userID,
		notificationType,
		notificationPayload,
	)
	if err != nil {
		return err
	}
	err = s.NotificationService.Publish(ctx, notification)
	if err != nil {
		return err
	}

	return nil
}

func (s *DuelService) sendNotificationBulk(
	ctx context.Context,
	notificationType uint8,
	notificationPayload []*model.UserNotificationPayload,
) error {
	for _, n := range notificationPayload {
		notification, err := model.NewNotification(
			n.UserID,
			notificationType,
			n.Notification,
		)
		if err != nil {
			return err
		}

		err = s.NotificationService.Publish(
			ctx,
			notification,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DuelService) sendVotedForNotification(
	ctx context.Context,
	userID uuid.UUID,
	duel *model.Duel,
	notificationPayload model.NotificationPayload,
) error {
	err := s.sendNotification(
		ctx,
		userID,
		model.NotificationVotedFor,
		notificationPayload,
	)
	if err != nil {
		return err
	}

	duel.PlayersCount++

	if duel.PlayersCount == 100 ||
		duel.PlayersCount == 500 ||
		duel.PlayersCount == 1000 {

		notification := &model.DuelPlayersJoinedNotification{
			DuelID:       duel.ID,
			DuelName:     duel.Question,
			PlayersCount: duel.PlayersCount,
		}

		err := s.sendNotification(
			ctx,
			duel.OwnerID,
			model.NotificationDuelPlayersJoined,
			notification,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DuelService) collectDuelRefundNotifications(
	ctx context.Context,
	duel *model.Duel,
) ([]*model.UserNotificationPayload, error) {
	var notifications []*model.UserNotificationPayload
	refundedPlayers, err := s.PlayerRepository.GetRefundedPlayersByID(ctx, duel.ID)
	if err != nil {
		return nil, err
	}

	for _, p := range refundedPlayers {
		notification := &model.DuelResolveNotification{
			DuelID:   duel.ID,
			DuelName: duel.Question,
			VotedFor: p.Answer,
			Amount:   duel.DuelPrice,
			Status:   model.StatusDuelRefund,
		}

		notifications = append(
			notifications,
			&model.UserNotificationPayload{
				UserID:       p.UserID,
				Notification: notification,
			},
		)
	}

	return notifications, nil
}

func (s *DuelService) sendNotificationsForDuelResolve(
	ctx context.Context,
	params *model.DuelResolveNotificationParams,
) error {
	var wrongAnswer uint8 = 0
	if params.DuelResolveParams.Answer == 0 {
		wrongAnswer = 1
	}

	var notifications []*model.UserNotificationPayload
	for _, id := range params.WinnerIDs {
		notification := &model.DuelResolveNotification{
			DuelID:   params.Duel.ID,
			DuelName: params.Duel.Question,
			VotedFor: params.DuelResolveParams.Answer,
			Amount:   params.WinAmount,
			Status:   model.StatusDuelWon,
		}

		notifications = append(
			notifications,
			&model.UserNotificationPayload{
				UserID:       id,
				Notification: notification,
			},
		)
	}

	loserIDs, err := s.PlayerRepository.GetDuelLosersIDs(
		ctx,
		params.Duel.ID,
		wrongAnswer,
		params.DuelResolveParams.JoinNotBefore,
	)
	if err != nil {
		return apperrors.Internal("failed to get duel loser ids", err)
	}

	for _, id := range loserIDs {
		notification := &model.DuelResolveNotification{
			DuelID:   params.Duel.ID,
			DuelName: params.Duel.Question,
			VotedFor: wrongAnswer,
			Amount:   params.Duel.DuelPrice,
			Status:   model.StatusDuelLost,
		}

		notifications = append(
			notifications,
			&model.UserNotificationPayload{
				UserID:       id,
				Notification: notification,
			},
		)
	}

	refundNotifications, err := s.collectDuelRefundNotifications(ctx, params.Duel)
	if err != nil {
		return err
	}
	notifications = append(notifications, refundNotifications...)

	if params.CreatorCommision != 0 {
		notification := &model.DuelResolveNotification{
			DuelID:   params.Duel.ID,
			DuelName: params.Duel.Question,
			Amount:   params.CreatorCommision,
			Status:   model.StatusDuelCommission,
		}
		notifications = append(
			notifications,
			&model.UserNotificationPayload{
				UserID:       params.Duel.OwnerID,
				Notification: notification,
			},
		)
	}

	return s.sendNotificationBulk(ctx, model.NotificationDuelResolve, notifications)
}

func (s *DuelService) sendDuelRefundNotification(
	ctx context.Context,
	duel *model.Duel,
) error {

	notifications, err := s.collectDuelRefundNotifications(ctx, duel)
	if err != nil {
		return err
	}

	err = s.sendNotificationBulk(context.Background(), model.NotificationDuelRefund, notifications)
	if err != nil {
		return err
	}

	return nil
}
