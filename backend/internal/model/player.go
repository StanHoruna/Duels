package model

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"math"
	"math/rand/v2"
	"time"
)

const (
	PlayerStatusActive   uint8 = 0
	PlayerStatusResolved uint8 = 1
	PlayerStatusRefunded uint8 = 2
)

type Player struct {
	bun.BaseModel `bun:"table:players,alias:players" json:"-"`

	ID          uuid.UUID `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	UserID      uuid.UUID `bun:"type:uuid" json:"user_id"`
	DuelID      uuid.UUID `bun:"type:uuid" json:"duel_id"`
	WinAmount   float64   `bun:"type:int" json:"win_amount"`
	Answer      uint8     `bun:"type:int" json:"answer"`
	FinalStatus uint8     `bun:"type:smallint" json:"final_status"`
	IsWinner    bool      `bun:"type:bool" json:"is_winner"`
	CreatedAt   time.Time `bun:",column:created_at,notnull,default:current_timestamp" json:"created_at"`
}

type PlayerWithAddress struct {
	bun.BaseModel `bun:"table:players,alias:players" json:"-"`

	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Answer        uint8     `json:"answer"`
	PublicAddress string    `json:"public_address"`
}

type PlayerShow struct {
	bun.BaseModel `bun:"table:players,alias:players" json:"-"`

	Player
	Username string `bun:",column:username,type:text" json:"username"`
	ImageUrl string `bun:",column:image_url,type:text" json:"image_url"`
}

func DuelByCreateReq(req *CreateDuelReq, user *User) *Duel {
	roomNum := uint64(rand.Int64N(math.MaxUint32-10_000) + 10_000 + 1)

	bgUrl := fmt.Sprintf("background_%d.svg", rand.Int64N(5))
	if len(req.BgURL) > 0 {
		bgUrl = req.BgURL
	}

	now := time.Now()
	return &Duel{
		ID:         uuid.New(),
		OwnerID:    user.ID,
		RoomNumber: roomNum,
		Username:   user.Username.String(),
		Status:     DuelStatusInProcess,
		ImageURL:   req.ImageURL,
		BgURL:      bgUrl,
		Question:   req.Question,
		DuelPrice:  req.DuelPrice,
		Commission: req.Commission,
		DuelInfo:   req.DuelInfo,
		EventDate:  req.EventDate,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
