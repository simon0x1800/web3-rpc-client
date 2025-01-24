package pyweb3

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Web3Client wraps ethclient.Client to provide Ethereum interaction capabilities
type Web3Client struct {
	client *ethclient.Client
}

// NewWeb3Client creates a new Web3Client instance
func NewWeb3Client(url string) (*Web3Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Web3Client{client: client}, nil
}

// SendTransaction sends an Ethereum transaction
func (w *Web3Client) SendTransaction(from, to common.Address, amount *big.Int) (*types.Transaction, error) {
	nonce, err := w.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, err
	}

	gasPrice, err := w.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, to, amount, 21000, gasPrice, nil)
	return tx, nil
}

// SendRawTransaction sends a signed transaction
func (w *Web3Client) SendRawTransaction(tx *types.Transaction) error {
	return w.client.SendTransaction(context.Background(), tx)
}
