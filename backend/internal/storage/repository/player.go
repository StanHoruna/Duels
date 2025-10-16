package repository

import (
	"context"
	"duels-api/internal/model"
	"duels-api/pkg/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PlayerRepository struct {
	repository.Generic[model.Player, uuid.UUID]
}

func NewPlayerRepository(
	genericRepository repository.Generic[model.Player, uuid.UUID],
) *PlayerRepository {
	return &PlayerRepository{
		Generic: genericRepository,
	}
}

func (r *PlayerRepository) WithTx(tx bun.Tx) *PlayerRepository {
	return &PlayerRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *PlayerRepository) UserAlreadyParticipant(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (bool, error) {
	player := new(model.Player)

	ok, err := r.DB.NewSelect().
		Model(player).
		Where("user_id = ? and duel_id = ?", userID, duelID).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *PlayerRepository) GetAllPlayersByDuelID(
	ctx context.Context,
	duelID uuid.UUID,
	options *repository.Options,
) ([]model.PlayerShow, error) {
	players := make([]model.PlayerShow, 0)

	q := r.DB.NewSelect().
		Model(&players).
		ColumnExpr("distinct players.id, players.user_id, players.duel_id, players.answer, players.created_at, u.username, u.image_url").
		Join("inner join users u on u.id = players.user_id").
		Where("players.duel_id = ?", duelID)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	correctAnswer uint8,
	deadline time.Time,
) ([]model.Player, error) {
	players := make([]model.Player, 0)

	err := r.DB.NewSelect().
		Model(&players).
		Where("duel_id = ?", duelID).
		Where("answer = ?", correctAnswer).
		Where("created_at <= ?", deadline).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelLosersIDs(
	ctx context.Context,
	duelID uuid.UUID,
	wrongAnswer uint8,
	deadline time.Time,
) (uuid.UUIDs, error) {
	loserIDs := make(uuid.UUIDs, 0)

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Column("user_id").
		Where("duel_id = ?", duelID).
		Where("answer = ?", wrongAnswer).
		Where("created_at <= ?", deadline).
		Scan(ctx, &loserIDs)
	if err != nil {
		return nil, err
	}
	fmt.Println(loserIDs)

	return loserIDs, nil
}

func (r *PlayerRepository) CountDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	correctAnswer uint8,
	deadline time.Time,
) (int, error) {
	return r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Where("answer = ?", correctAnswer).
		Where("created_at <= ?", deadline).
		Count(ctx)
}

func (r *PlayerRepository) UpdateDuelWinners(
	ctx context.Context,
	winners []model.Player,
	winAmount float64,
) error {
	ids := make(uuid.UUIDs, 0, len(winners))
	for _, winner := range winners {
		ids = append(ids, winner.ID)
	}

	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("players.id IN (?)", bun.In(ids)).
		Set("is_winner = ?", true).
		Set("win_amount = ?", winAmount).
		Set("final_status = ?", model.PlayerStatusResolved).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) SetStatusToAll(
	ctx context.Context,
	duelID uuid.UUID,
	status uint8,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("players.duel_id = ?", duelID).
		Set("final_status = ?", status).
		Exec(ctx)

	return err
}

// SetStatus — упрощённый сеттер статуса для набора игроков с адресами (выплаты/рефанды)
func (r *PlayerRepository) SetStatus(
	ctx context.Context,
	players []model.PlayerWithAddress,
	status uint8,
) error {
	data := r.DB.NewValues(&players)

	_, err := r.DB.NewUpdate().
		With("_data", data).
		Model((*model.Player)(nil)).
		TableExpr("_data").
		Where("players.id = _data.id").
		Set("final_status = ?", status).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) MarkPlayersAsRefunded(
	ctx context.Context,
	duelID uuid.UUID,
	joinNotBefore time.Time,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Where("created_at > ?", joinNotBefore).
		Set("final_status = ?", model.PlayerStatusRefunded).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (*model.Player, error) {
	player := new(model.Player)

	err := r.DB.NewSelect().
		Model(player).
		Where("user_id = ?", userID).
		Where("duel_id = ?", duelID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return player, nil
}

// --- crypto helpers (адреса берём из users) ---

func (r *PlayerRepository) GetCryptoDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	answer uint8,
) ([]model.PlayerWithAddress, error) {
	players := make([]model.PlayerWithAddress, 0)

	err := r.DB.NewSelect().
		Model(&players).
		ColumnExpr("players.id").
		ColumnExpr("players.user_id").
		ColumnExpr("players.answer").
		ColumnExpr("u.public_address").
		Join("left join users AS u on players.user_id = u.id").
		Where("players.duel_id = ?", duelID).
		Where("players.answer = ?", answer).
		Where("players.final_status = ?", model.PlayerStatusActive).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetCryptoDuelPlayers(
	ctx context.Context,
	duelID uuid.UUID,
) ([]model.PlayerWithAddress, error) {
	players := make([]model.PlayerWithAddress, 0)

	err := r.DB.NewSelect().
		Model(&players).
		ColumnExpr("players.id").
		ColumnExpr("players.user_id").
		ColumnExpr("players.answer").
		ColumnExpr("u.public_address").
		Join("left join users AS u on players.user_id = u.id").
		Where("players.duel_id = ?", duelID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelPlayersToRefund(
	ctx context.Context,
	duelID uuid.UUID,
	votedAfter time.Time,
) ([]model.PlayerWithAddress, error) {
	players := make([]model.PlayerWithAddress, 0)

	err := r.DB.NewSelect().
		Model(&players).
		ColumnExpr("players.id").
		ColumnExpr("players.user_id").
		ColumnExpr("players.answer").
		ColumnExpr("u.public_address").
		Join("left join users AS u on players.user_id = u.id").
		Where("players.duel_id = ?", duelID).
		Where("players.final_status = ?", model.PlayerStatusActive).
		Where("players.created_at > ?", votedAfter).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) CountDuelPlayersToRefund(
	ctx context.Context,
	duelID uuid.UUID,
	deadline time.Time,
) (int, error) {
	return r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Where("final_status = ?", model.PlayerStatusActive).
		Where("created_at > ?", deadline).
		Count(ctx)
}

func (r *PlayerRepository) HasUserJoinedMoreThanOneDuel(ctx context.Context, userID uuid.UUID) (bool, error) {
	var duelCount int

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT duel_id)").
		Where("user_id = ?", userID).
		Scan(ctx, &duelCount)
	if err != nil {
		return false, err
	}

	return duelCount > 1, nil
}

func (r *PlayerRepository) GetRefundedPlayersByID(
	ctx context.Context,
	duelID uuid.UUID,
) ([]*model.Player, error) {
	var players []*model.Player

	err := r.DB.NewSelect().
		Model(&players).
		Where("duel_id = ?", duelID).
		Where("final_status = ?", model.PlayerStatusRefunded).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}
