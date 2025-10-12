package service

import (
	"context"
	"duels-api/config"
	sol "duels-api/internal/client/solana"
	"duels-api/internal/model"
	"duels-api/internal/storage/repository"
	"duels-api/pkg/apperrors"
	"duels-api/pkg/sigtracker"
	"encoding/base64"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-resty/resty/v2"
)

const (
	USDCPriceMultiplier        = 1_000_000
	Finalized                  = rpc.CommitmentFinalized
	Confirmed                  = rpc.CommitmentConfirmed
	TxConfirmationTimeout      = 30 * time.Second
	CUExtraCapacityCoefficient = 1.20
	FallBackCUTransfer         = uint32(25_000)
)

var ZeroValuePublicKey solana.PublicKey
var TransactionMaxRetryCount uint = 10

type WalletService struct {
	SolanaRPC    *rpc.Client
	HTTPClient   *resty.Client
	TxRepository *repository.TransactionRepository

	SigTracker      *sigtracker.TxTracker
	PriorityTracker *PriorityTracker

	solanaAdminPrivateKey solana.PrivateKey
	contractAddress       string
	contractAddressAPI    string

	usdcMintAddress solana.PublicKey
}

func NewWalletService(
	c *config.Config,
	solanaRPC *rpc.Client,
	txRepo *repository.TransactionRepository,
	sigTracker *sigtracker.TxTracker,
	priorityTracker *PriorityTracker,
) (*WalletService, error) {
	solanaAdminPrivateKey, err := solana.PrivateKeyFromBase58(c.App.SolanaAdminPrivateKey)
	if err != nil {
		return nil, apperrors.Internal("failed to get solana contract admin private key")
	}

	usdcMintAddress, err := solana.PublicKeyFromBase58(c.App.USDCMintAddress)
	if err != nil {
		return nil, apperrors.Internal("failed to parse usdc mint address", err)
	}

	return &WalletService{
		SolanaRPC:             solanaRPC,
		HTTPClient:            resty.New(),
		TxRepository:          txRepo,
		SigTracker:            sigTracker,
		PriorityTracker:       priorityTracker,
		solanaAdminPrivateKey: solanaAdminPrivateKey,
		contractAddress:       c.App.ContractAddress,
		contractAddressAPI:    c.App.ContractAddressApi,
		usdcMintAddress:       usdcMintAddress,
	}, nil
}

func (s *WalletService) validateCreateCryptoDuelSCTransaction(
	ctx context.Context,
	duel *model.CreateDuelReq,
) (uint64, error) {
	sig, err := solana.SignatureFromBase58(duel.Hash)
	if err != nil {
		return 0, apperrors.BadRequest("failed to parse tx hash")
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return 0, apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return 0, apperrors.Internal("tx was not confirmed: "+sig.String(), nil)
	}

	txInfo, err := s.SolanaRPC.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return 0, apperrors.Internal("get transaction info", err)
	}

	logs := strings.Join(txInfo.Meta.LogMessages, ", ")

	if !strings.Contains(logs, "Instruction: Init") {
		return 0, apperrors.BadRequest("invalid instruction")
	}
	if !strings.Contains(logs, fmt.Sprintf("Description: %s", duel.Question)) {
		return 0, apperrors.BadRequest("invalid description")
	}
	if !strings.Contains(logs, fmt.Sprintf("Bet: %d USDC", int64(duel.DuelPrice))) {
		return 0, apperrors.BadRequest("invalid bet")
	}
	if !strings.Contains(logs, fmt.Sprintf("Current executing program address: %s", s.contractAddress)) {
		return 0, apperrors.BadRequest("invalid program address")
	}

	var roomNumber uint64

	match := model.JoinedRoomRegex.FindStringSubmatch(logs)

	if len(match) > 1 {
		roomNumber, err = strconv.ParseUint(match[1], 10, 64)
	} else {
		err = apperrors.BadRequest("room number not found")
	}
	if err != nil {
		return 0, err
	}

	return roomNumber, nil
}

func (s *WalletService) InitAndJoinSolanaRoomWithExternalWallet(
	ctx context.Context,
	duel *model.Duel,
	user *model.User,
	answer uint8,
) (string, error) {
	duelPrice := duel.DuelPrice * USDCPriceMultiplier

	reqBody := map[string]interface{}{
		"description": duel.Question,
		"percent":     uint32(duel.Commission),
		"bet":         uint32(duelPrice),
		"pda_nr":      uint32(duel.RoomNumber),
	}

	txInit, err := s.GetTxFromContractService(reqBody, "init")
	if err != nil {
		return "", err
	}

	publicKey, err := solana.PublicKeyFromBase58(user.PublicAddress)
	if err != nil {
		return "", apperrors.Internal("failed to parse user's public key", err)
	}

	userTokenAccount, _, err := solana.FindAssociatedTokenAddress(publicKey, s.usdcMintAddress)
	if err != nil {
		return "", apperrors.Internal("failed to get user associated token address", err)
	}

	hasEnoughBalance, err := s.HasEnoughTokenBalance(ctx, userTokenAccount, duelPrice)
	if err != nil {
		return "", err
	}

	if !hasEnoughBalance {
		return "", apperrors.BadRequest("not enough balance to proceed a transaction")
	}

	reqBody = map[string]interface{}{
		"multiplier": 1,
		"answer":     answer,
		"pda_nr":     duel.RoomNumber,
		"payer":      publicKey,
	}

	txJoin, err := s.GetTxFromContractService(reqBody, "join")
	if err != nil {
		return "", err
	}

	recentBlockhashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return "", apperrors.Internal("failed to get recent blockhash", err)
	}

	instructions, err := GetTxInstructions(txInit, txJoin)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockhashResp.Value.Blockhash,
		solana.TransactionPayer(publicKey))
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to generate a transaction", err)
	}

	_, err = tx.PartialSign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(s.solanaAdminPrivateKey.PublicKey()) {
			return &s.solanaAdminPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to sign transaction", err)
	}

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", apperrors.ServiceUnavailable("failed to marshal transaction", err)
	}

	encodedTx := base64.StdEncoding.EncodeToString(txBytes)

	return encodedTx, nil
}

func (s *WalletService) validateJoinCryptoDuelSCTransaction(
	ctx context.Context,
	txHash string,
) error {
	sig, err := solana.SignatureFromBase58(txHash)
	if err != nil {
		return apperrors.BadRequest("failed to parse tx hash")
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	tx, err := s.SolanaRPC.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return apperrors.Internal("failed to get transaction by tx hash", err)
	}

	logs := strings.Join(tx.Meta.LogMessages, ", ")

	if !strings.Contains(logs, "Instruction: Join") {
		return apperrors.BadRequest("invalid instruction")
	}
	if !strings.Contains(logs, fmt.Sprintf("Current executing program address: %s", s.contractAddress)) {
		return apperrors.BadRequest("invalid program address")
	}

	return nil
}

func (s *WalletService) TransferBulkSolanaChain(
	ctx context.Context,
	amount uint64,
	players []model.PlayerWithAddress,
	mint solana.PublicKey,
) ([]string, error) {

	adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(s.solanaAdminPrivateKey.PublicKey(), mint)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	allTransferInstructions, err := getTransferInstruction(s.solanaAdminPrivateKey, amount, players, mint)
	if err != nil {
		return nil, err
	}

	separatedInstructions := separateInstructions(allTransferInstructions)
	txHashes := make([]string, 0, len(separatedInstructions))

	for _, instructions := range separatedInstructions {
		tx, err := s.NewTransactionForSimulation(
			instructions,
			txSignerPrivateKeyGetter(s.solanaAdminPrivateKey),
			solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
		if err != nil {
			return nil, err
		}

		computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
		if err != nil {
			if len(instructions) > 1 {
				computeUnits = FallBackCUTransfer * uint32(len(instructions))
			}
		}

		computeUnits = uint32(float64(computeUnits)*CUExtraCapacityCoefficient + 300)
		cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
			SetMicroLamports(s.PriorityTracker.GetHighPriorityMicroLamports()).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit price", err)
		}

		cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
			SetUnits(computeUnits).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
		}

		// Compute Unit Price and Compute Unit Limit instructions must be first
		instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

		tx, err = solana.NewTransaction(
			instructions,
			solana.Hash{},
			solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
		if err != nil {
			return nil, apperrors.Internal("failed to create transaction", err)
		}

		txHash, err := s.sendTransaction(
			ctx,
			tx,
			txSignerPrivateKeyGetter(s.solanaAdminPrivateKey))
		if err != nil {
			return nil, err
		}

		txHashes = append(txHashes, txHash.String())
	}

	return txHashes, nil
}

func (s *WalletService) CloseSolanaRoom(
	ctx context.Context,
	roomNumber uint64,
) (string, error) {
	adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(s.solanaAdminPrivateKey.PublicKey(), s.usdcMintAddress)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return "", apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	reqBody := map[string]any{
		"pda_nr": roomNumber,
	}

	instructions, err := s.GetInstructionsFromContractService(
		ctx,
		reqBody,
		"close",
		s.solanaAdminPrivateKey)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		instructions,
		solana.Hash{},
		solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
	if err != nil {
		return "", apperrors.Internal("failed to create transaction", err)
	}

	txHash, err := s.sendTxWithTracker(
		ctx,
		tx,
		txSignerPrivateKeyGetter(s.solanaAdminPrivateKey))
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}

func (s *WalletService) RewardDuelWinners(
	ctx context.Context,
	winAmount uint64,
	winners []model.PlayerWithAddress,
	mint solana.PublicKey,
) ([]string, error) {
	if len(winners) == 0 {
		return []string{}, nil
	}

	adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(s.solanaAdminPrivateKey.PublicKey(), mint)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	allTransferInstructions, err := getTransferInstruction(s.solanaAdminPrivateKey, winAmount, winners, mint)
	if err != nil {
		return nil, err
	}

	separatedInstructions := separateInstructions(allTransferInstructions)

	txHashes := make([]string, 0, len(separatedInstructions))

	for _, instructions := range separatedInstructions {
		tx, err := s.NewTransactionForSimulation(
			instructions,
			txSignerPrivateKeyGetter(s.solanaAdminPrivateKey),
			solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
		if err != nil {
			return nil, err
		}

		computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
		if err != nil {
			if len(instructions) > 1 {
				computeUnits = FallBackCUTransfer * uint32(len(instructions))
			}
		}

		computeUnits = uint32(float64(computeUnits)*CUExtraCapacityCoefficient + 300)
		cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
			SetMicroLamports(s.PriorityTracker.GetHighPriorityMicroLamports()).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit price", err)
		}

		cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
			SetUnits(computeUnits).
			ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
		}

		// Compute Unit Price and Compute Unit Limit instructions must be first
		instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

		tx, err = solana.NewTransaction(
			instructions,
			solana.Hash{},
			solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
		if err != nil {
			return nil, apperrors.Internal("failed to create transaction", err)
		}

		txHash, err := s.sendTransaction(
			ctx,
			tx,
			txSignerPrivateKeyGetter(s.solanaAdminPrivateKey))
		if err != nil {
			return nil, err
		}

		txHashes = append(txHashes, txHash.String())
	}

	return txHashes, nil
}

func (s *WalletService) RewardDuelOwnerWithCommission(
	ctx context.Context,
	publicAddress string,
	commissionReward uint64,
	mint solana.PublicKey,
) (string, error) {
	if commissionReward == 0 {
		return "", nil
	}

	adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(s.solanaAdminPrivateKey.PublicKey(), mint)
	if err != nil || adminTokenAccount == ZeroValuePublicKey {
		return "", apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	duelOwnerAddress, err := solana.PublicKeyFromBase58(publicAddress)
	if err != nil || duelOwnerAddress == ZeroValuePublicKey {
		return "", apperrors.BadRequest("recipient is not valid solana address", err)
	}

	duelOwnerATA, _, err := solana.FindAssociatedTokenAddress(duelOwnerAddress, mint)
	if err != nil || duelOwnerATA == ZeroValuePublicKey {
		return "", apperrors.BadRequest("failed to find recipient associated token account", err)
	}

	info, err := s.SolanaRPC.GetAccountInfo(ctx, duelOwnerATA)
	if err != nil && !errors.Is(err, rpc.ErrNotFound) {
		return "", apperrors.ServiceUnavailable("failed to get recipient's account info", err)
	}

	inst := make([]solana.Instruction, 0, 2)
	if info == nil || info.Value == nil || info.Value.Owner == ZeroValuePublicKey {
		initTokenAccountInstruction, err := associatedtokenaccount.NewCreateInstruction(
			s.solanaAdminPrivateKey.PublicKey(),
			duelOwnerAddress,
			mint).ValidateAndBuild()
		if err != nil {
			return "", apperrors.Internal("failed to build token account initialization instruction", err)
		}

		inst = append(inst, initTokenAccountInstruction)
	}

	transferInstruction, err := token.NewTransferInstruction(
		commissionReward,
		adminTokenAccount,
		duelOwnerATA,
		s.solanaAdminPrivateKey.PublicKey(),
		[]solana.PublicKey{s.solanaAdminPrivateKey.PublicKey()}).ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to build transfer transaction", err)
	}

	inst = append(inst, transferInstruction)

	txHash, err := s.SendTransaction(ctx, inst)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func (s *WalletService) SendTransaction(
	ctx context.Context,
	instructions []solana.Instruction,
) (string, error) {
	tx, err := s.NewTransactionForSimulation(
		instructions,
		txSignerPrivateKeyGetter(s.solanaAdminPrivateKey),
		solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
	if err != nil {
		return "", err
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		if len(instructions) > 1 {
			computeUnits = FallBackCUTransfer * uint32(len(instructions))
		}
	}

	computeUnits = uint32(float64(computeUnits)*CUExtraCapacityCoefficient + 300)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetHighPriorityMicroLamports()).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(computeUnits).
		ValidateAndBuild()
	if err != nil {
		return "", apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	// Compute Unit Price and Compute Unit Limit instructions must be first
	instructions = append([]solana.Instruction{cuPriceInstruction, cuLimitInstruction}, instructions...)

	tx, err = solana.NewTransaction(
		instructions,
		solana.Hash{},
		solana.TransactionPayer(s.solanaAdminPrivateKey.PublicKey()))
	if err != nil {
		return "", apperrors.Internal("failed to create transaction", err)
	}

	txHash, err := s.sendTransaction(
		ctx,
		tx,
		txSignerPrivateKeyGetter(s.solanaAdminPrivateKey))
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}

func (s *WalletService) NewTransactionForSimulation(
	instructions []solana.Instruction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
	opts ...solana.TransactionOption,
) (*solana.Transaction, error) {
	tx, err := solana.NewTransaction(
		instructions,
		solana.Hash{}, // latest block hash will be set just before transaction sending or transaction simulation
		opts...)
	if err != nil {
		return nil, apperrors.Internal("failed to create transaction for simulation", err)
	}

	_, err = tx.Sign(privateKeyGetter)
	if err != nil {
		return nil, apperrors.Internal("failed to sign a transaction for simulation", err)
	}

	return tx, nil
}

func txSignerPrivateKeyGetter(sender solana.PrivateKey) func(solana.PublicKey) *solana.PrivateKey {
	return func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(sender.PublicKey()) {
			return &sender
		}

		return nil
	}
}

func (s *WalletService) GetSimulationComputeUnits(
	ctx context.Context,
	tx *solana.Transaction,
) (uint32, error) {
	opts := &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
		SigVerify:              false, // conflicts with ReplaceRecentBlockhash
	}

	result, err := s.simulateTransaction(ctx, tx, opts)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, apperrors.Internal("transaction simulation: tx value is nil")
	}

	if result.Err != nil {
		zap.L().Error("transaction simulation", zap.Any("err", result.Err))

		if s.isAccountUninitialized(fmt.Errorf("%v", result.Err)) {
			return 0, apperrors.Internal(model.ErrSolanaAccountUninitialized)
		}

		return 0, sol.ParseLogsForError(result.Logs)
	}

	if result.UnitsConsumed == nil {
		return 0, apperrors.Internal("transaction simulation: 0 units consumed")
	}

	return uint32(*result.UnitsConsumed), nil
}

func (s *WalletService) simulateTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	opts *rpc.SimulateTransactionOpts,
) (*rpc.SimulateTransactionResult, error) {
	sTx, err := s.SolanaRPC.SimulateTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to send simulation transaction", err)
	}

	if sTx == nil {
		return nil, apperrors.ServiceUnavailable("failed to get simulation transaction compute units, tx is nil", err)
	}

	return sTx.Value, nil
}

func (s *WalletService) sendTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
) (solana.Signature, error) {
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: Finalized,
		MaxRetries:          &TransactionMaxRetryCount,
	}

	if err := s.RefreshBlockHash(ctx, &tx.Message); err != nil {
		return solana.Signature{}, err
	}

	_, err := tx.Sign(privateKeyGetter)
	if err != nil {
		return solana.Signature{}, apperrors.Internal("failed to sign transaction", err)
	}

	sig, err := s.SolanaRPC.SendTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		return solana.Signature{}, apperrors.Internal("failed to send transaction", err)
	}

	return sig, nil
}

func (s *WalletService) RefreshBlockHash(ctx context.Context, txMessage *solana.Message) error {
	recentBlockHashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return apperrors.ServiceUnavailable("failed to get latest block hash", err)
	}

	txMessage.RecentBlockhash = recentBlockHashResp.Value.Blockhash
	return nil
}

func (s *WalletService) sendTxWithTracker(
	ctx context.Context,
	tx *solana.Transaction,
	privateKeyGetter func(key solana.PublicKey) *solana.PrivateKey,
) (solana.Signature, error) {
	sig, err := s.sendTransaction(ctx, tx, privateKeyGetter)
	if err != nil {
		return solana.Signature{}, err
	}

	sent, err := s.SigTracker.SubscribeForSignatureStatus(sig, TxConfirmationTimeout)
	if err != nil && !errors.Is(err, rpc.ErrNotConfirmed) {
		return solana.Signature{}, apperrors.Internal("subscribe to signature status", err)
	}

	if !sent {
		return solana.Signature{}, apperrors.Internal("tx was not confirmed: " + sig.String())
	}

	return sig, nil
}

const TransferInstructionsPerTransaction = 32

func separateInstructions(instructions []solana.Instruction) [][]solana.Instruction {
	separatedInstructions := make([][]solana.Instruction, 0, len(instructions)/TransferInstructionsPerTransaction+1)

	for i := 0; i < len(instructions); i += TransferInstructionsPerTransaction {
		sliceEnd := min(i+TransferInstructionsPerTransaction, len(instructions))
		separatedInstructions = append(separatedInstructions, instructions[i:sliceEnd])
	}

	return separatedInstructions
}

func getTransferInstruction(
	sender solana.PrivateKey,
	amount uint64,
	players []model.PlayerWithAddress,
	mint solana.PublicKey,
) ([]solana.Instruction, error) {
	senderTokenAccount, _, err := solana.FindAssociatedTokenAddress(sender.PublicKey(), mint)
	if err != nil || senderTokenAccount == ZeroValuePublicKey {
		return nil, apperrors.Internal("failed to find associated token account for user rewarding", err)
	}

	instructions := make([]solana.Instruction, 0, len(players)+2)

	for _, player := range players {
		recipient, err := solana.PublicKeyFromBase58(player.PublicAddress)
		if err != nil || recipient == ZeroValuePublicKey {
			return nil, apperrors.BadRequest("recipient is not valid solana address", err)
		}

		recipientTokenAccount, _, err := solana.FindAssociatedTokenAddress(recipient, mint)
		if err != nil || recipientTokenAccount == ZeroValuePublicKey {
			return nil, apperrors.BadRequest("failed to find recipient associated token account", err)
		}

		transferInstruction, err := token.NewTransferInstruction(
			amount,
			senderTokenAccount,
			recipientTokenAccount,
			sender.PublicKey(),
			[]solana.PublicKey{sender.PublicKey()}).ValidateAndBuild()
		if err != nil {
			return nil, apperrors.Internal("failed to build transfer transaction", err)
		}

		instructions = append(instructions, transferInstruction)
	}

	return instructions, nil
}

func (s *WalletService) GetInstructionsFromContractService(
	ctx context.Context,
	reqBody map[string]any,
	endpoint string,
	signer solana.PrivateKey,
) ([]solana.Instruction, error) {
	tx, err := s.GetTxFromContractService(reqBody, endpoint)
	if err != nil {
		return nil, err
	}

	recentBlockHashResp, err := s.SolanaRPC.GetLatestBlockhash(ctx, Finalized)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to get latest block hash", err)
	}

	tx.Message.RecentBlockhash = recentBlockHashResp.Value.Blockhash

	_, err = tx.Sign(txSignerPrivateKeyGetter(signer))
	if err != nil {
		return nil, apperrors.Internal("failed to sign a transaction", err)
	}

	computeUnits, err := s.GetSimulationComputeUnits(ctx, tx)
	if err != nil {
		return nil, err
	}

	computeUnits = uint32(float64(computeUnits) * CUExtraCapacityCoefficient)
	cuPriceInstruction, err := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(s.PriorityTracker.GetMediumPriorityMicroLamports()).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit price", err)
	}

	cuLimitInstruction, err := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(computeUnits).
		ValidateAndBuild()
	if err != nil {
		return nil, apperrors.Internal("failed to set transaction compute unit limit", err)
	}

	instructions := make([]solana.Instruction, 0, 3)
	instructions = append(instructions, cuPriceInstruction, cuLimitInstruction)

	for _, instruction := range tx.Message.Instructions {
		inst, err := decompileInstruction(instruction, tx)
		if err != nil {
			return nil, apperrors.ServiceUnavailable("failed to decompile instruction", err)
		}

		instructions = append(instructions, inst)
	}

	return instructions, nil
}

func (s *WalletService) GetTxFromContractService(
	reqBody map[string]any,
	endpoint string,
) (*solana.Transaction, error) {
	resp, err := s.HTTPClient.R().
		SetBody(reqBody).
		SetHeader("Content-Type", "application/json").
		Put(s.contractAddressAPI + endpoint)
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to call contract service", err)
	}

	tx, err := ExtractTxFromResp(resp)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func ExtractTxFromResp(resp *resty.Response) (*solana.Transaction, error) {
	if resp == nil || !resp.IsSuccess() {
		return nil, apperrors.ServiceUnavailable("failed to get transaction from contract service: resp is nil")
	}

	var respData model.ContractServiceTxResp
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return nil, apperrors.ServiceUnavailable("failed to unmarshal raw transaction", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(respData.RawTx))
	if err != nil {
		return nil, apperrors.ServiceUnavailable("failed to decode a transaction", err)
	}

	return tx, nil
}

func (s *WalletService) HasEnoughTokenBalance(
	ctx context.Context,
	ata solana.PublicKey,
	requiredAmount float64,
) (bool, error) {
	tokenBalance, err := s.getTokenBalance(ctx, ata, Confirmed)
	if err != nil {
		return false, err
	}

	return float64(tokenBalance) >= requiredAmount, nil
}

func (s *WalletService) getTokenBalance(
	ctx context.Context,
	ata solana.PublicKey,
	commitment rpc.CommitmentType,
) (uint64, error) {
	balance, err := s.SolanaRPC.GetTokenAccountBalance(ctx, ata, commitment)
	if err != nil {
		if s.isAccountUninitialized(err) {
			return 0, model.ErrTokenAccountUninitialized
		}

		return 0, apperrors.ServiceUnavailable("failed to get ata balance", err)
	}

	if balance == nil || balance.Value == nil {
		return 0, apperrors.ServiceUnavailable("failed to get ata balance: balance is nil", nil)
	}

	balanceAmount, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
	if err != nil {
		return 0, apperrors.ServiceUnavailable("failed to parse token balance amount", err)
	}

	return balanceAmount, nil
}

func (s *WalletService) isAccountUninitialized(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "could not find account") ||
		strings.Contains(err.Error(), "AccountNotFound")
}

func GetTxInstructions(txs ...*solana.Transaction) ([]solana.Instruction, error) {
	instructionsCount := 0
	for _, tx := range txs {
		if tx == nil {
			continue
		}

		instructionsCount += len(tx.Message.Instructions)
	}

	instructions := make([]solana.Instruction, 0, instructionsCount)

	for _, tx := range txs {
		if tx == nil {
			continue
		}

		for _, instruction := range tx.Message.Instructions {
			inst, err := decompileInstruction(instruction, tx)
			if err != nil {
				return nil, apperrors.ServiceUnavailable("failed to decompile instruction", err)
			}

			instructions = append(instructions, inst)
		}
	}

	return instructions, nil
}

type CompiledInstruction struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func (c *CompiledInstruction) ProgramID() solana.PublicKey {
	return c.programID
}

func (c *CompiledInstruction) Accounts() []*solana.AccountMeta {
	return c.accounts
}

func (c *CompiledInstruction) Data() ([]byte, error) {
	return c.data, nil
}

func decompileInstruction(instruction solana.CompiledInstruction, tx *solana.Transaction) (*CompiledInstruction, error) {
	programID, err := tx.ResolveProgramIDIndex(instruction.ProgramIDIndex)
	if err != nil {
		return nil, fmt.Errorf("resolve program ID: %w", err)
	}

	accounts, err := instruction.ResolveInstructionAccounts(&tx.Message)
	if err != nil {
		return nil, fmt.Errorf("resolve instruction account: %w", err)
	}

	return &CompiledInstruction{
		programID: programID,
		accounts:  accounts,
		data:      instruction.Data,
	}, nil
}
