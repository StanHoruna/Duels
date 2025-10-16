package model

import (
	"duels-api/pkg/apperrors"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	NotificationDuelResolve uint8 = iota + 10
	NotificationDuelRefund

	NotificationVotedFor

	NotificationDuelModeration

	NotificationDuelPlayersJoined

	NotificationDuelEndingSoon
)

const (
	StatusDuelWon uint8 = iota
	StatusDuelLost
	StatusDuelRefund
	StatusDuelCommission
)

type NotificationPayload interface {
	Marshal() ([]byte, error)
}

type Notification struct {
	bun.BaseModel `bun:"table:notifications,alias:n" json:"-"`

	ID               uuid.UUID       `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	UserID           uuid.UUID       `bun:"user_id,type:uuid,notnull" json:"user_id"`
	NotificationType uint8           `bun:"type:integer,notnull" json:"notification_type"`
	Data             json.RawMessage `bun:"type:jsonb" json:"data"`
	IsRead           bool            `bun:"type:,notnull,default:false" json:"is_read"`
	CreatedAt        time.Time       `bun:",notnull,default:current_timestamp" json:"created_at"`
}

func NewNotification(
	userID uuid.UUID,
	notificationType uint8,
	notification NotificationPayload,
) (*Notification, error) {
	data, err := notification.Marshal()
	if err != nil {
		return nil, apperrors.Internal("failed to marshal notification", err)
	}

	return &Notification{
		ID:               uuid.New(),
		UserID:           userID,
		NotificationType: notificationType,
		Data:             data,
		IsRead:           false,
		CreatedAt:        time.Now(),
	}, nil
}

func (n *Notification) MarshalBinary() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Notification) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, n)
}

type UserNotificationPayload struct {
	UserID       uuid.UUID
	Notification NotificationPayload
}

type DuelResolveNotification struct {
	DuelID   uuid.UUID `json:"duel_id"`
	DuelName string    `json:"duel_name"`
	VotedFor uint8     `json:"voted_for"`
	Amount   float64   `json:"amount"`
	Status   uint8     `json:"status"`
}

func (n *DuelResolveNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

type VotedForNotification struct {
	DuelID   uuid.UUID `json:"duel_id"`
	DuelName string    `json:"duel_name"`
	VotedFor uint8     `json:"voted_for"`
}

func (n *VotedForNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

type DuelModerationNotification struct {
	DuelID             uuid.UUID `json:"duel_id"`
	DuelName           string    `json:"duel_name"`
	IsApproved         bool      `json:"is_approved"`
	CancellationReason string    `json:"cancellation_reason"`
}

func (n *DuelModerationNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

type DuelPlayersJoinedNotification struct {
	DuelID       uuid.UUID `json:"duel_id"`
	DuelName     string    `json:"duel_name"`
	PlayersCount uint64    `json:"players_count"`
}

func (n *DuelPlayersJoinedNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

type DuelEndingSoonNotification struct {
	DuelID   uuid.UUID `json:"duel_id"`
	DuelName string    `json:"duel_name"`
	Deadline uint64    `json:"deadline"`
}

func (n *DuelEndingSoonNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

type DuelResolveNotificationParams struct {
	WinnerIDs         []uuid.UUID
	Duel              *Duel
	DuelResolveParams *DuelResolveParams
	WinAmount         float64
	CreatorCommision  float64
}
