package service

import (
	"context"
	"duels-api/config"
	"duels-api/internal/model"
	"duels-api/internal/storage/repository"
	"duels-api/pkg/apperrors"
	repo "duels-api/pkg/repository"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type DuelService struct {
	WalletService       *WalletService
	UserRepository      *repository.UserRepository
	TxRepository        *repository.TransactionRepository
	DuelRepository      *repository.DuelRepository
	PlayerRepository    *repository.PlayerRepository
	HTTPShareClient     *resty.Client
	TransactionManager  *repo.TransactionManager
	ShareImageAPI       string
	USDCMintAddress     solana.PublicKey
	USDCMintDecimals    uint8
	NotificationService *NotificationService
}

func NewDuelService(
	c *config.Config,
	walletService *WalletService,
	userRepository *repository.UserRepository,
	txRepository *repository.TransactionRepository,
	duelRepository *repository.DuelRepository,
	playerRepository *repository.PlayerRepository,
	transactionManager *repo.TransactionManager,
	notificationService *NotificationService,
) (*DuelService, error) {
	USDCMintAddress, err := solana.PublicKeyFromBase58(c.App.USDCMintAddress)
	if err != nil {
		return nil, apperrors.Internal("invalid USDC mint address", err)
	}
	return &DuelService{
		WalletService:       walletService,
		UserRepository:      userRepository,
		TxRepository:        txRepository,
		DuelRepository:      duelRepository,
		PlayerRepository:    playerRepository,
		HTTPShareClient:     resty.New(),
		TransactionManager:  transactionManager,
		ShareImageAPI:       c.App.ShareImageAPI,
		USDCMintAddress:     USDCMintAddress,
		USDCMintDecimals:    c.App.USDCMintDecimals,
		NotificationService: notificationService,
	}, nil
}

func (s *DuelService) GetAllDuelsUnauthorized(ctx context.Context, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetAllDuels(ctx, uuid.Nil, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels", err)
	}

	return duels, nil
}

func (s *DuelService) GetAllDuels(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetAllDuels(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get duels", err)
	}

	return duels, nil
}

func (s *DuelService) CountAllDuels(ctx context.Context, options *repo.Options) (int, error) {
	count, err := s.DuelRepository.CountWithOptions(ctx, options)
	if err != nil {
		return 0, apperrors.Internal("failed to count all duels", err)
	}

	return count, nil
}

func (s *DuelService) GetDuelByIDUnauthorized(ctx context.Context, duelID uuid.UUID) (*model.DuelShow, error) {
	duel, err := s.DuelRepository.GetDuelShowByID(ctx, uuid.Nil, duelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel", err)
	}

	return duel, nil
}

func (s *DuelService) GetMyDuels(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetUserDuels(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get my duels", err)
	}

	return duels, nil
}

func (s *DuelService) GetMyDuelsAsParticipant(ctx context.Context, userID uuid.UUID, options *repo.Options) ([]model.DuelShow, error) {
	duels, err := s.DuelRepository.GetUserDuelsAsParticipant(ctx, userID, options)
	if err != nil {
		return nil, apperrors.Internal("failed to get my duels", err)
	}

	return duels, nil
}

func (s *DuelService) GetDuelByID(ctx context.Context,
	duelID uuid.UUID,
	userID uuid.UUID,
) (*model.DuelShow, []model.PlayerShow, error) {
	duel, err := s.DuelRepository.GetDuelShowByID(ctx, userID, duelID)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get duel", err)
	}

	players, err := s.PlayerRepository.GetAllPlayersByDuelID(ctx, duelID, nil)
	if err != nil {
		return nil, nil, apperrors.Internal("failed to get players", err)
	}

	return duel, players, nil
}

func (s *DuelService) CreateCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.CreateDuelReq,
) (*model.CreateCryptoDuelResp, error) {
	roomNumber, err := s.WalletService.
		validateCreateCryptoDuelSCTransaction(
			ctx,
			req,
		)
	if err != nil {
		zap.L().Warn("transaction validation failed", zap.Error(err))
		return nil, apperrors.BadRequest("transaction validation failed")
	}

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	duel := model.DuelByCreateReq(req, user)

	duel.RoomNumber = roomNumber

	if err = s.createAndJoinCryptoDuel(ctx, duel, user, req.Answer, req.Hash); err != nil {
		return nil, err
	}

	if err = s.sendDuelShareImageReq(duel); err != nil {
		zap.L().Warn("share image request failed", zap.Error(err))
	}

	return &model.CreateCryptoDuelResp{
		Duel: duel,
	}, nil
}

func (s *DuelService) createAndJoinCryptoDuel(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	ownerAnswer uint8,
	txHash string,
) error {
	join := &model.JoinDuelReq{
		DuelID: duel.ID,
		Answer: ownerAnswer,
	}

	err := s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			err := s.DuelRepository.WithTx(tx).Create(ctx, duel)
			if err != nil {
				return apperrors.Internal("failed to create duel", err)
			}

			_, err = s.DuelRepository.WithTx(tx).JoinDuel(ctx, user.ID, join, duel)
			if err != nil {
				return apperrors.Internal("failed to join owner to duel", err)
			}

			if txHash != "" {
				txRecord := &model.TransactionType{
					Signature: txHash,
					TxType:    model.TransactionTypeDuelPrediction,
				}

				if err = s.TxRepository.WithTx(tx).Create(ctx, txRecord); err != nil {
					return apperrors.Internal("failed to create transaction record", err)
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	notification := &model.VotedForNotification{
		DuelID:   duel.ID,
		DuelName: duel.Question,
		VotedFor: ownerAnswer,
	}

	err = s.sendNotification(ctx, user.ID, model.NotificationVotedFor, notification)
	if err != nil {
		zap.L().Error("failed to send notification", zap.Error(err))
	}

	return nil
}

func (s *DuelService) SignCreateCryptoDuelTransaction(
	ctx context.Context,
	userID uuid.UUID,
	req *model.CreateDuelReq,
) (string, error) {
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return "", apperrors.Internal("failed to get user", err)
	}

	duel := model.DuelByCreateReq(req, user)

	duel.Status = model.DuelStatusInProcess

	tx, err := s.WalletService.InitAndJoinSolanaRoomWithExternalWallet(ctx, duel, user, req.Answer)
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (s *DuelService) JoinExternalWalletCryptoDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
) (*model.JoinCryptoDuelResp, error) {
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel", err)
	}

	if err = s.isAbleToJoinDuel(ctx, duel, user.ID); err != nil {
		return nil, err
	}

	err = s.WalletService.
		validateJoinCryptoDuelSCTransaction(
			ctx,
			req.Hash,
		)
	if err != nil {
		zap.L().Warn("transaction validation error", zap.Error(err))
		return nil, apperrors.BadRequest("transaction validation failed")
	}

	var player *model.Player
	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			player, err = s.DuelRepository.WithTx(tx).JoinDuel(ctx, userID, req, duel)
			if err != nil {
				return apperrors.Internal("failed to join duel", err)
			}

			txRecord := &model.TransactionType{
				Signature: req.Hash,
				TxType:    model.TransactionTypeDuelPrediction,
			}
			if err = s.TxRepository.WithTx(tx).Create(ctx, txRecord); err != nil {
				return apperrors.Internal("failed to create transaction record", err)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	notification := &model.VotedForNotification{
		DuelID:   duel.ID,
		DuelName: duel.Question,
		VotedFor: req.Answer,
	}

	err = s.sendVotedForNotification(ctx, user.ID, duel, notification)
	if err != nil {
		zap.L().Error("failed to send notification", zap.Error(err))
	}

	return &model.JoinCryptoDuelResp{Player: player}, nil

}

func (s *DuelService) SignJoinCryptoDuelTransaction(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
) (string, error) {
	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return "", apperrors.Internal("failed to get user", err)
	}

	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return "", apperrors.Internal("failed to get duel", err)
	}

	if err = s.isAbleToJoinDuel(ctx, duel, userID); err != nil {
		return "", err
	}

	tx, err := s.WalletService.JoinSolanaRoomWithExternalWallet(ctx, duel, user, req.Answer)
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (s *DuelService) ResolveCryptoDuelByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
	req *model.DuelResolveReq,
) ([]string, error) {
	duel, err := s.DuelRepository.GetByID(ctx, req.DuelID)
	if err != nil {
		return nil, apperrors.Internal("failed to get duel by id", err)
	}

	err = s.validateResolveDuelByOwner(ownerID, duel)
	if err != nil {
		return nil, err
	}

	params := &model.DuelResolveParams{
		DuelID:        duel.ID,
		Answer:        req.Answer,
		JoinNotBefore: duel.EventDate,
	}

	txHashes, err := s.resolveCryptoDuel(ctx, duel, params)
	if err != nil {
		return nil, err
	}

	return txHashes, nil
}

func (s *DuelService) resolveCryptoDuel(
	ctx context.Context,
	duel *model.Duel,
	req *model.DuelResolveParams,
) ([]string, error) {
	duelWinners, err := s.PlayerRepository.GetDuelWinners(ctx, duel.ID, req.Answer, req.JoinNotBefore)
	if err != nil {
		return nil, apperrors.Internal("failed to count players with specific answer", err)
	}
	allDuelWinnersCount := uint64(len(duelWinners))

	playersToRefund, err := s.PlayerRepository.CountDuelPlayersToRefund(ctx, duel.ID, req.JoinNotBefore)
	if err != nil {
		return nil, apperrors.Internal("failed to count players that must be refunded", err)
	}

	playersPool := duel.PlayersCount - uint64(playersToRefund)
	if playersPool == 0 || playersPool == allDuelWinnersCount || allDuelWinnersCount == 0 {
		return s.cancelCryptoDuel(ctx, duel, model.AutoCancelReq(duel))
	}

	var refundedPlayersTxHashes []string
	if playersToRefund > 0 {
		refundedPlayersTxHashes, err = s.partialCryptoRefund(ctx, duel, req.JoinNotBefore)
		if err != nil {
			return nil, err
		}
	}

	unpaidWinners, err := s.PlayerRepository.GetCryptoDuelWinners(ctx, duel.ID, req.Answer)
	if err != nil {
		return nil, apperrors.Internal("failed to count players with specific answer", err)
	}

	if len(unpaidWinners) == 0 {
		return []string{}, nil
	}

	playersCount := duel.PlayersCount - duel.RefundedPlayersCount
	duelParams := model.NewDuelParams(
		duel.DuelPrice, duel.Commission,
		playersCount, allDuelWinnersCount,
	)

	var (
		duelRewardTxHashes []string
		priceMultiplier    = float64(USDCPriceMultiplier)
		winAmount          = duelParams.CalculateFinalCryptoReward(priceMultiplier)
		playersWinAmount   = float64(winAmount) / priceMultiplier
		mint               = s.USDCMintAddress
	)

	duelRewardTxHashes, err = s.WalletService.RewardDuelWinners(ctx, winAmount, unpaidWinners, mint)
	if err != nil {
		return nil, err
	}

	txRecords := make([]model.TransactionType, 0, len(duelRewardTxHashes)+len(refundedPlayersTxHashes)+1)

	txRecords = append(
		txRecords,
		model.NewTransactionsWithSameType(
			model.TransactionTypeDuelReward,
			duelRewardTxHashes...,
		)...,
	)

	if len(refundedPlayersTxHashes) > 0 {
		txRecords = append(
			txRecords,
			model.NewTransactionsWithSameType(
				model.TransactionTypeDuelRefund,
				refundedPlayersTxHashes...,
			)...,
		)
	}

	var (
		txHash            string
		commissionRewards = &model.CommissionRewards{}
	)

	commissionRewards, err = s.rewardWithCommissions(ctx, duel, &duelParams, mint)
	if err != nil {
		return nil, err
	}

	txRecords = append(txRecords, commissionRewards.TXRecords...)

	txHash, err = s.WalletService.CloseSolanaRoom(ctx, duel.RoomNumber)
	if err != nil {
		zap.L().Error("failed to close solana room after resolve",
			zap.Error(err),
			zap.String("duel_id", duel.ID.String()),
			zap.Uint64("room_number", duel.RoomNumber))
	}

	duel.Status = model.DuelStatusResolved
	duel.FinalResult = &req.Answer
	duel.WinnersCount = allDuelWinnersCount

	winnerIDs := make([]uuid.UUID, 0, len(duelWinners))
	for i := range duelWinners {
		winnerIDs = append(winnerIDs, duelWinners[i].UserID)
	}

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			err = s.PlayerRepository.WithTx(tx).UpdateDuelWinners(ctx, duelWinners, playersWinAmount)
			if err != nil {
				return err
			}

			if err = s.DuelRepository.WithTx(tx).Update(ctx, duel); err != nil {
				return err
			}

			return s.TxRepository.WithTx(tx).BulkInsert(ctx, txRecords)
		})
	if err != nil {
		return nil, apperrors.Internal("failed to resolve a duel", err)
	}

	allTxHashes := append(duelRewardTxHashes, commissionRewards.CreatorCommissionTxHash, txHash)
	allTxHashes = append(allTxHashes, refundedPlayersTxHashes...)

	go func() {
		err = s.sendNotificationsForDuelResolve(
			context.Background(),
			&model.DuelResolveNotificationParams{
				WinnerIDs:         winnerIDs,
				Duel:              duel,
				DuelResolveParams: req,
				WinAmount:         float64(winAmount),
				CreatorCommision:  float64(commissionRewards.CreatorCommissionReward),
			},
		)
		if err != nil {
			zap.L().Error("failed to send notification", zap.Error(err))
		}
	}()

	return allTxHashes, nil
}

func (s *DuelService) rewardWithCommissions(
	ctx context.Context,
	duel *model.Duel,
	duelParams *model.DuelParams,
	mint solana.PublicKey,
) (*model.CommissionRewards, error) {
	duelOwner, err := s.UserRepository.GetByID(ctx, duel.OwnerID)
	if err != nil {
		return &model.CommissionRewards{}, apperrors.Internal("failed to get duel owner by id", err)
	}

	creatorCommissionReward := duelParams.CalculateCryptoCommissionReward(USDCPriceMultiplier)

	creatorCommissionTxHash, err := s.WalletService.RewardDuelOwnerWithCommission(
		ctx,
		duelOwner.PublicAddress,
		creatorCommissionReward,
		mint,
	)
	if err != nil {
		return &model.CommissionRewards{}, err
	}

	var txRecords []model.TransactionType

	if creatorCommissionTxHash != "" {
		txRecords = append(
			txRecords,
			model.NewTransactionsWithSameType(
				model.TransactionTypeDuelCommission,
				creatorCommissionTxHash,
			)...,
		)
	}

	return &model.CommissionRewards{
		TXRecords:               txRecords,
		CreatorCommissionTxHash: creatorCommissionTxHash,
		CreatorCommissionReward: creatorCommissionReward,
	}, nil
}

func (s *DuelService) cancelCryptoDuel(
	ctx context.Context,
	duel *model.Duel,
	req *model.DuelCancelReq,
) ([]string, error) {
	var (
		txHashes          = make([]string, 0)
		roomClosingTxHash = ""
	)

	if s.hasChargedDuelPriceFromUser(duel.PlayersCount, duel.Status, req.Status) {
		players, err := s.PlayerRepository.GetCryptoDuelPlayers(ctx, duel.ID)
		if err != nil {
			return nil, apperrors.Internal("failed to get duel players", err)
		}
		var (
			duelPrice = uint64(duel.DuelPrice * USDCPriceMultiplier)
			mint      = s.USDCMintAddress
		)
		txHashes, err = s.WalletService.TransferBulkSolanaChain(ctx, duelPrice, players, mint)
		if err != nil {
			return nil, apperrors.ServiceUnavailable("failed to refund duel: "+duel.ID.String(), err)
		}

		roomClosingTxHash, err = s.WalletService.CloseSolanaRoom(ctx, duel.RoomNumber)
		if err != nil {
			zap.L().Error("failed to close solana room after refund",
				zap.Error(err),
				zap.String("duel_id", duel.ID.String()),
				zap.Uint64("room_number", duel.RoomNumber))
		}
		go func() {
			err = s.sendDuelRefundNotification(context.Background(), duel)
			if err != nil {
				zap.L().Error("failed to send notification", zap.Error(err))
			}
		}()
	}

	duel.Status = req.Status
	duel.CancellationReason = req.CancellationReason

	err := s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			err := s.DuelRepository.Update(ctx, duel)
			if err != nil {
				return apperrors.Internal("failed to update duel status", err)
			}

			err = s.PlayerRepository.SetStatusToAll(ctx, duel.ID, model.PlayerStatusRefunded)
			if err != nil {
				return apperrors.Internal("failed to update crypto players status", err)
			}

			if len(txHashes) > 0 {
				err = s.TxRepository.WithTx(tx).BulkInsertWithSameTxType(
					ctx,
					model.TransactionTypeDuelRefund,
					txHashes)
				if err != nil {
					return apperrors.Internal("failed to create transaction records", err)
				}
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	return append(txHashes, roomClosingTxHash), nil
}

func (s *DuelService) partialCryptoRefund(
	ctx context.Context,
	duel *model.Duel,
	votedAfter time.Time,
) ([]string, error) {
	if duel.Status != model.DuelStatusInProcess {
		return []string{}, nil
	}

	players, err := s.PlayerRepository.GetDuelPlayersToRefund(ctx, duel.ID, votedAfter)
	if err != nil {
		return nil, apperrors.Internal("failed to get crypto duel players", err)
	}

	var (
		duelPrice = uint64(duel.DuelPrice * USDCPriceMultiplier)
		mint      = s.USDCMintAddress
	)

	txHashes, err := s.WalletService.TransferBulkSolanaChain(ctx, duelPrice, players, mint)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to refund duel: "+duel.ID.String(), err)
	}

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			err = s.PlayerRepository.WithTx(tx).SetStatus(ctx, players, model.PlayerStatusRefunded)
			if err != nil {
				return apperrors.Internal("failed to update crypto players status", err)
			}

			duel.RefundedPlayersCount += uint64(len(players))
			if err = s.DuelRepository.WithTx(tx).Update(ctx, duel); err != nil {
				return apperrors.Internal("failed to update duel", err)
			}

			err = s.TxRepository.WithTx(tx).BulkInsertWithSameTxType(
				ctx,
				model.TransactionTypeDuelRefund,
				txHashes)
			if err != nil {
				return apperrors.Internal("failed to create transaction records", err)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	return txHashes, nil
}
