// Package txbatch is used to batch transactions for a main chain node
package txbatch

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/brainbot-com/shutter/shuttermint/medley"
	"github.com/brainbot-com/shutter/shuttermint/sandbox"
)

type TXBatch struct {
	Ethclient    *ethclient.Client
	TransactOpts *bind.TransactOpts

	key          *ecdsa.PrivateKey
	transactions []*types.Transaction
}

func New(ctx context.Context, client *ethclient.Client, key *ecdsa.PrivateKey) (*TXBatch, error) {
	opts, err := sandbox.InitTransactOpts(ctx, client, key)
	if err != nil {
		return nil, err
	}
	return &TXBatch{
		Ethclient:    client,
		TransactOpts: opts,
		key:          key,
		transactions: nil,
	}, nil
}

func (txbatch *TXBatch) Add(tx *types.Transaction) {
	txbatch.transactions = append(txbatch.transactions, tx)
	txbatch.TransactOpts.Nonce.SetInt64(txbatch.TransactOpts.Nonce.Int64() + 1)
}

func (txbatch *TXBatch) WaitMined(ctx context.Context) ([]*types.Receipt, error) {
	txHashes := []common.Hash{}
	for _, tx := range txbatch.transactions {
		txHashes = append(txHashes, tx.Hash())
	}

	return medley.WaitMinedMany(ctx, txbatch.Ethclient, txHashes)
}