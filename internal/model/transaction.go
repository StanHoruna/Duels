package model

import (
	"github.com/uptrace/bun"
)

const (
	TransactionTypeDuelPrediction uint8 = 1
	TransactionTypeDuelRefund     uint8 = 2
	TransactionTypeDuelCommission uint8 = 3
	TransactionTypeDuelReward     uint8 = 4
)

type TransactionType struct {
	bun.BaseModel `bun:"table:transactions,alias:tx" json:"-"`

	Signature string `bun:"signature,pk,type:CHAR(88)" json:"signature"`
	TxType    uint8  `bun:"tx_type,type:SMALLINT,notnull" json:"tx_type"`
}

func NewTransaction(txType uint8, signature string) TransactionType {
	return TransactionType{Signature: signature, TxType: txType}
}

func NewTransactionsWithSameType(txType uint8, signatures ...string) []TransactionType {
	txs := make([]TransactionType, 0, len(signatures))
	for _, s := range signatures {
		txs = append(txs, NewTransaction(txType, s))
	}
	return txs
}

type GetTransactionTypesReq struct {
	Signatures []string `json:"signatures"`
}
