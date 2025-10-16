package repository

import (
	"context"
	"database/sql"
	"duels-api/internal/model"
	"duels-api/pkg/repository"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type DuelRepository struct {
	repository.Generic[model.Duel, uuid.UUID]
}

func NewDuelRepository(
	genericRepository repository.Generic[model.Duel, uuid.UUID],
) *DuelRepository {
	return &DuelRepository{
		Generic: genericRepository,
	}
}

func (r *DuelRepository) WithTx(tx bun.Tx) *DuelRepository {
	return &DuelRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *DuelRepository) CreateBulk(ctx context.Context, duels []model.Duel) error {
	_, err := r.DB.NewInsert().Model(&duels).Exec(ctx)
	return err
}

func (r *DuelRepository) GetAllDuels(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id")
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetAllDuelsWhereParticipate(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("inner join players p on p.duel_id = duels.id and p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id")
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetDuelShowByID(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (*model.DuelShow, error) {
	duel := new(model.DuelShow)

	q := r.DB.NewSelect().
		Model(duel).
		ColumnExpr("duels.*").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.id = ?", duelID)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duel, nil
}

func (r *DuelRepository) GetHistoryByUserID(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer as your_answer").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("p.user_id = ? and duels.status in (?, ?, ?, ?)",
			userID,
			model.DuelStatusAutoCancelled,
			model.DuelStatusAdminCancelled,
			model.DuelStatusResolved,
			model.DuelStatusRefund,
		)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetUserDuels(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("(p.user_id IS NOT NULL) as joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer as your_answer").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.owner_id = ?", userID)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) JoinDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
	duel *model.Duel,
) (*model.Player, error) {
	player := &model.Player{
		ID:        uuid.New(),
		UserID:    userID,
		DuelID:    req.DuelID,
		Answer:    req.Answer,
		CreatedAt: time.Now(),
	}
	_, err := r.DB.NewInsert().Model(player).Exec(ctx)
	if err != nil {
		return nil, err
	}

	_, err = r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("players_count = players_count + 1").
		Set("status = ?", duel.Status).
		Where("id = ?", req.DuelID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (r *DuelRepository) JoinDuelBulk(
	ctx context.Context,
	userID uuid.UUID,
	duels []model.Duel,
	answers []uint8,
) error {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("error caught", zap.Any("err", err))
		}
	}()
	if len(duels) != len(answers) {
		return fmt.Errorf("number of duels does not match number of answers")
	}

	players := make([]model.Player, 0, len(duels))
	for i, duel := range duels {
		player := model.Player{
			ID:        uuid.New(),
			UserID:    userID,
			DuelID:    duel.ID,
			Answer:    answers[i],
			CreatedAt: time.Now(),
		}
		players = append(players, player)
	}

	_, err := r.DB.NewInsert().Model(&players).Exec(ctx)
	if err != nil {
		return err
	}

	values := r.DB.NewValues(&players)

	_, err = r.DB.NewUpdate().
		With("_data", values).
		Model((*model.Duel)(nil)).
		TableExpr("_data").
		Set("players_count = players_count + 1").
		Where("duels.id = _data.duel_id").
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *DuelRepository) FindOwnerIDByDuelID(
	ctx context.Context,
	duelID uuid.UUID,
) (uuid.UUID, error) {
	var ownerID uuid.UUID

	err := r.DB.NewSelect().
		Table("duels").
		Column("owner_id").
		Where("id = ?", duelID).
		Scan(ctx, &ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}

	return ownerID, nil
}
