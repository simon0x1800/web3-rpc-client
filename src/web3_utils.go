package pyweb3

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionWatcher handles transaction monitoring
type TransactionWatcher struct {
	client  *Web3Client
	timeout time.Duration
	blocks  uint64
}

// NewTransactionWatcher creates a new transaction watcher
func NewTransactionWatcher(client *Web3Client, timeout time.Duration, blocks uint64) *TransactionWatcher {
	return &TransactionWatcher{
		client:  client,
		timeout: timeout,
		blocks:  blocks,
	}
}

// WaitForConfirmations waits for a specific number of block confirmations
func (tw *TransactionWatcher) WaitForConfirmations(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, tw.timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for confirmations")
		case <-ticker.C:
			receipt, err := tw.client.client.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue
			}

			currentBlock, err := tw.client.client.BlockNumber(ctx)
			if err != nil {
				continue
			}

			confirmations := currentBlock - receipt.BlockNumber.Uint64()
			if confirmations >= tw.blocks {
				return receipt, nil
			}
		}
	}
}

// GasEstimator handles gas estimation with safety margins
type GasEstimator struct {
	client *Web3Client
	margin uint64
}

// NewGasEstimator creates a new gas estimator
func NewGasEstimator(client *Web3Client, marginPercent uint64) *GasEstimator {
	return &GasEstimator{
		client: client,
		margin: marginPercent,
	}
}

// EstimateGasWithMargin estimates gas with a safety margin
func (ge *GasEstimator) EstimateGasWithMargin(msg ethereum.CallMsg) (uint64, error) {
	gas, err := ge.client.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}

	margin := gas * ge.margin / 100
	return gas + margin, nil
}

// GetOptimalGasPrice suggests an optimal gas price based on recent blocks
func (ge *GasEstimator) GetOptimalGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := ge.client.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	margin := new(big.Int).Mul(gasPrice, big.NewInt(int64(ge.margin)))
	margin = margin.Div(margin, big.NewInt(100))

	return new(big.Int).Add(gasPrice, margin), nil
}
