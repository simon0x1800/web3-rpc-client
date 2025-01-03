package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Web3Client struct {
	client *ethclient.Client
	ctx    context.Context
}

// NewWeb3Client creates a new Web3 client instance
func NewWeb3Client(url string) (*Web3Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}

	return &Web3Client{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// GetLatestBlock retrieves the latest block from the blockchain
func (w *Web3Client) GetLatestBlock() (*types.Block, error) {
	block, err := w.client.BlockByNumber(w.ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest block: %v", err)
	}
	return block, nil
}

// WaitForTransaction waits for a transaction to be mined and returns its receipt
func (w *Web3Client) WaitForTransaction(txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(w.ctx, timeout)
	defer cancel()

	receipt, err := w.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %v", err)
	}
	return receipt, nil
}

// EstimateGas estimates the gas needed for a transaction
func (w *Web3Client) EstimateGas(to common.Address, data []byte) (uint64, error) {
	gas, err := w.client.EstimateGas(w.ctx, ethereum.CallMsg{
		To:   &to,
		Data: data,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}
	return gas, nil
}

// WatchEvents listens for specific events from a contract
func (w *Web3Client) WatchEvents(contractAddress common.Address, eventSignature string, fromBlock *big.Int) (chan types.Log, ethereum.Subscription, error) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{common.HexToHash(eventSignature)}},
		FromBlock: fromBlock,
	}

	logs := make(chan types.Log)
	sub, err := w.client.SubscribeFilterLogs(w.ctx, query, logs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to event logs: %v", err)
	}

	return logs, sub, nil
}

// GetNonce gets the next nonce for an address
func (w *Web3Client) GetNonce(address common.Address) (uint64, error) {
	nonce, err := w.client.PendingNonceAt(w.ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %v", err)
	}
	return nonce, nil
}

// CallContract performs a contract call without creating a transaction
func (w *Web3Client) CallContract(contractAddress common.Address, data []byte) ([]byte, error) {
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := w.client.CallContract(w.ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %v", err)
	}
	return result, nil
}

// GetLogs retrieves logs matching the given criteria
func (w *Web3Client) GetLogs(contractAddress common.Address, topics []common.Hash, fromBlock *big.Int, toBlock *big.Int) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{topics},
	}

	logs, err := w.client.FilterLogs(w.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	return logs, nil
}

// GetTransactionByHash retrieves a transaction by its hash
func (w *Web3Client) GetTransactionByHash(txHash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := w.client.TransactionByHash(w.ctx, txHash)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get transaction: %v", err)
	}
	return tx, isPending, nil
}

// GetContractEvents decodes contract events using the provided ABI
func (w *Web3Client) GetContractEvents(contractAbi abi.ABI, logs []types.Log) ([]interface{}, error) {
	var events []interface{}
	for _, log := range logs {
		for _, event := range contractAbi.Events {
			if log.Topics[0] == event.ID {
				var decoded interface{}
				err := contractAbi.UnpackIntoInterface(&decoded, event.Name, log.Data)
				if err != nil {
					return nil, fmt.Errorf("failed to decode event: %v", err)
				}
				events = append(events, decoded)
			}
		}
	}
	return events, nil
}

// Close closes the client connection
func (w *Web3Client) Close() {
	w.client.Close()
}

// Example usage
func main() {
	client, err := NewWeb3Client("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Get latest block
	block, err := client.GetLatestBlock()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Latest block number: %d\n", block.Number())
}
