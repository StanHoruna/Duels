package solana

import (
	"duels-api/config"
	"github.com/gagliardetto/solana-go/rpc"
)

func NewClient(c *config.Config) *rpc.Client {
	return rpc.New(c.App.SolanaNodeURL)
}
