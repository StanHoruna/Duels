package solana

import (
	"duels-api/pkg/apperrors"
	"strings"
)

var (
	ErrInsufficientFunds = apperrors.PaymentRequired("insufficient funds for proceeding a transaction")
)

func ParseLogsForError(logs []string) error {
	for _, log := range logs {
		switch {
		case IsInsufficientFundForCommission(log):
			return ErrInsufficientFunds
		}
	}

	return apperrors.Internal("transaction: result err")
}

func IsInsufficientFundForCommission(log string) bool {
	return strings.Contains(log, "insufficient")
}
