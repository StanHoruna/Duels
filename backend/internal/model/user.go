package model

import (
	"duels-api/pkg/mtype"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID            uuid.UUID      `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Username      mtype.Username `bun:",type:varchar(17),notnull" json:"username"`
	ImageUrl      string         `bun:",type:varchar(100)" json:"image_url"`
	PublicAddress string         `bun:",type:varchar(100),unique,nullzero" json:"public_address"`
	CreatedAt     time.Time      `bun:",notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time      `bun:",notnull,default:current_timestamp" json:"updated_at"`
}

func NewUser(
	username mtype.Username,
	profileImageURL string,
) *User {
	now := time.Now().UTC()
	return &User{
		ID:        uuid.New(),
		Username:  username,
		ImageUrl:  profileImageURL,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type AuthWithWallet struct {
	Address string `json:"address" binding:"required"`
	Secret  string `json:"secret" binding:"required"`
}

type SignInJWTResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type UsernameChange struct {
	bun.BaseModel `bun:"table:users,alias:u" json:"-"`

	Username mtype.Username `bun:"username" json:"username"`
}

type UserStats struct {
	Participated   uint64  `bun:"participated" json:"participated"`
	WinsCount      uint64  `bun:"wins_count" json:"wins_count"`
	LossesCount    uint64  `bun:"losses_count" json:"losses_count"`
	RefundedCount  uint64  `bun:"refunded_count" json:"refunded_count"`
	EarnedAmount   float64 `bun:"earned_amount" json:"earned_amount"`
	LostAmount     float64 `bun:"lost_amount" json:"lost_amount"`
	RefundedAmount float64 `bun:"refunded_amount" json:"refunded_amount"`
	NetProfit      float64 `bun:"net_profit" json:"net_profit"`
}
