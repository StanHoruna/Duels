package model

import "duels-api/pkg/apperrors"

var (
	ErrTokenAccountUninitialized  = apperrors.NotFound("token account is not initialized")
	ErrSolanaAccountUninitialized = "solana account is not initialized"
)

type ContractServiceTxResp struct {
	RawTx []byte `json:"raw_tx"`
}

type CommissionRewards struct {
	TXRecords               []TransactionType
	CreatorCommissionTxHash string
	CreatorCommissionReward uint64
}
