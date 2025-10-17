package model

import (
	"math"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	_ = uint8(iota)
	DuelStatusInReview
	DuelStatusAutoCancelled
	DuelStatusAdminCancelled
	DuelStatusInProcess
	DuelStatusResolved
	DuelStatusRefund
)

const (
	USDCDuelMinJoinPrice = 1.0
	USDCDuelMaxJoinPrice = 5000.0
)

const (
	SamePredictionCancellationReason     = "All users made the same prediction"
	LackOfParticipantsCancellationReason = "The duel was canceled due to a lack of participants"
)

type Duel struct {
	bun.BaseModel `bun:"table:duels,alias:duels" json:"-"`

	ID                   uuid.UUID `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	OwnerID              uuid.UUID `bun:"owner_id,type:uuid,notnull" json:"owner_id"`
	RoomNumber           uint64    `bun:"room_number,type:integer,nullzero" json:"room_number"` // todo bigint
	PlayersCount         uint64    `bun:"players_count,type:integer,notnull,default:0" json:"players_count"`
	RefundedPlayersCount uint64    `bun:"refunded_players_count,type:integer,notnull,default:0" json:"refunded_players_count"`
	WinnersCount         uint64    `bun:"winners_count,type:integer,notnull,default:0" json:"winners_count"`

	Username string `bun:"username,type:varchar(17),notnull" json:"username"`

	Status   uint8  `bun:"status,type:integer,notnull,default:0" json:"status"`
	ImageURL string `bun:"image_url,type:text" json:"image_url"`
	BgURL    string `bun:"bg_url,type:text" json:"bg_url"`

	Question   string         `bun:"question,type:text" json:"question"`
	DuelPrice  float64        `bun:"duel_price,type:int,notnull" json:"duel_price"`
	Commission uint64         `bun:"commission,type:integer,notnull" json:"commission"`
	DuelInfo   map[string]any `bun:"duel_info,type:json" json:"duel_info"`
	EventDate  time.Time      `bun:"event_date,notnull,default:current_timestamp" json:"event_date"`

	FinalResult        *uint8 `bun:"final_result,type:integer" json:"final_result"`
	CancellationReason string `bun:"cancellation_reason,type:text" json:"cancellation_reason"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

var JoinedRoomRegex = regexp.MustCompile(`joined room (\d+)!`)

type CryptoDuelInfo struct {
	ID    int     `json:"coin_id"`
	Price float64 `json:"target_price"`
	Type  int     `json:"direction"`
}

func (c *CryptoDuelInfo) DetermineWinningBet(coinPrice float64) uint8 {
	if c.Type == 0 && coinPrice <= c.Price {
		return 1
	}
	if c.Type == 1 && coinPrice >= c.Price {
		return 1
	}
	return 0
}

func GetCryptoDuelInfo(duelInfo map[string]any) (*CryptoDuelInfo, bool) {
	if duelInfo == nil {
		return nil, false
	}
	id, ok := duelInfo["coin_id"].(float64)
	if !ok {
		return nil, false
	}
	price, ok := duelInfo["target_price"].(float64)
	if !ok {
		return nil, false
	}
	dir, ok := duelInfo["direction"].(float64)
	if !ok {
		return nil, false
	}
	return &CryptoDuelInfo{ID: int(id), Price: price, Type: int(dir)}, true
}

type DuelShow struct {
	bun.BaseModel `bun:"table:duels,alias:duels" json:"-"`

	Duel
	OwnerImageURL string `bun:"owner_image_url" json:"owner_image_url"`
	YesCount      uint64 `bun:",column:yes_count" json:"yes_count"`
	NoCount       uint64 `bun:",column:no_count" json:"no_count"`
	Joined        bool   `bun:",column:joined" json:"joined"`
	YourAnswer    *int   `bun:",column:your_answer" json:"your_answer"`
	PlayerStatus  uint8  `bun:",column:player_status" json:"player_status"`
}

type CreateDuelReq struct {
	ImageURL string `json:"image_url"`
	BgURL    string `json:"bg_url"`
	Question string `json:"question"`

	DuelPrice  float64        `json:"duel_price"`
	Commission uint64         `json:"commission"`
	DuelInfo   map[string]any `json:"duel_info"`
	EventDate  time.Time      `json:"event_date"`

	Answer uint8 `json:"answer"`

	Hash string `json:"tx_hash"`
}

type JoinDuelReq struct {
	DuelID         uuid.UUID `json:"duel_id"`
	Answer         uint8     `json:"answer"`
	InvitedBy      string    `json:"invited_by"`
	ExternalSource string    `json:"external_source"`
	Hash           string    `json:"tx_hash"`
}

type DuelResolveReq struct {
	DuelID uuid.UUID `json:"duel_id"`
	Answer uint8     `json:"answer"`
}

type DuelResolveParams struct {
	DuelID        uuid.UUID `json:"duel_id"`
	Answer        uint8     `json:"answer"`
	JoinNotBefore time.Time `json:"not_before"`
}

type DuelCancelReq struct {
	DuelID             uuid.UUID `json:"duel_id"`
	Status             uint8     `json:"status"`
	CancellationReason string    `json:"cancellation_reason"`
}

type DuelParams struct {
	Pool         float64
	Commission   float64
	PlayersCount float64
	WinnersCount float64
}

func NewDuelParams(price float64, commission uint64, playersCount, winnersCount uint64) DuelParams {
	return DuelParams{
		Pool:         float64(playersCount) * price,
		Commission:   float64(commission),
		PlayersCount: float64(playersCount),
		WinnersCount: float64(winnersCount),
	}
}
func (p DuelParams) CalculateFinalReward() float64 {
	percentValue := math.Round((p.Pool / 100) * p.Commission)
	finalPool := math.Floor(p.Pool - percentValue)
	return math.Floor(finalPool / p.WinnersCount)
}
func (p DuelParams) CalculateFinalCryptoReward(priceMultiplier float64) uint64 {
	percentValue := p.Pool * p.Commission * priceMultiplier / 100
	finalPool := p.Pool*priceMultiplier - percentValue
	return uint64(finalPool / p.WinnersCount)
}
func (p DuelParams) CalculateCryptoCommissionReward(priceMultiplier float64) uint64 {
	percentValue := p.Pool * p.Commission * priceMultiplier / 100
	return uint64(percentValue / 2)
}

type JoinSolanaRoomResp struct {
	TxHash string `json:"tx_hash"`
}

type CreateCryptoDuelResp struct {
	Duel   *Duel               `json:"duel"`
	Result *JoinSolanaRoomResp `json:"result"`
}

type JoinCryptoDuelResp struct {
	Player *Player             `json:"player"`
	Result *JoinSolanaRoomResp `json:"result"`
}

func AutoCancelReq(duel *Duel) *DuelCancelReq {
	cancellationReason := SamePredictionCancellationReason
	if duel.PlayersCount <= 1 {
		cancellationReason = LackOfParticipantsCancellationReason
	}

	return &DuelCancelReq{
		DuelID:             duel.ID,
		Status:             DuelStatusRefund,
		CancellationReason: cancellationReason,
	}
}
