package main

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWeb3Client is a mock implementation of the Web3Client
type MockWeb3Client struct {
	mock.Mock
}

func (m *MockWeb3Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockWeb3Client) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func TestNewTransactionWatcher(t *testing.T) {
	client := &MockWeb3Client{}
	timeout := 10 * time.Second
	blocks := uint64(6)

	tw := NewTransactionWatcher(client, timeout, blocks)

	assert.NotNil(t, tw)
	assert.Equal(t, client, tw.client)
	assert.Equal(t, timeout, tw.timeout)
	assert.Equal(t, blocks, tw.blocks)
}

func TestWaitForConfirmations_Success(t *testing.T) {
	client := &MockWeb3Client{}
	txHash := common.HexToHash("0x12345")
	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
	}

	client.On("TransactionReceipt", mock.Anything, txHash).Return(receipt, nil).Times(1)
	client.On("BlockNumber", mock.Anything).Return(uint64(106), nil).Times(1)

	tw := NewTransactionWatcher(client, 10*time.Second, 6)
	ctx := context.Background()

	result, err := tw.WaitForConfirmations(ctx, txHash)

	assert.NoError(t, err)
	assert.Equal(t, receipt, result)
	client.AssertExpectations(t)
}

func TestWaitForConfirmations_Timeout(t *testing.T) {
	client := &MockWeb3Client{}
	txHash := common.HexToHash("0x12345")

	client.On("TransactionReceipt", mock.Anything, txHash).Return(nil, errors.New("not found")).Maybe()
	client.On("BlockNumber", mock.Anything).Return(uint64(0), nil).Maybe()

	tw := NewTransactionWatcher(client, 2*time.Second, 6)
	ctx := context.Background()

	result, err := tw.WaitForConfirmations(ctx, txHash)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for confirmations")
	client.AssertExpectations(t)
}

func TestWaitForConfirmations_ClientError(t *testing.T) {
	client := &MockWeb3Client{}
	txHash := common.HexToHash("0x12345")

	client.On("TransactionReceipt", mock.Anything, txHash).Return(nil, errors.New("client error")).Times(3)
	client.On("BlockNumber", mock.Anything).Return(uint64(0), errors.New("client error")).Times(3)

	tw := NewTransactionWatcher(client, 5*time.Second, 6)
	ctx := context.Background()

	result, err := tw.WaitForConfirmations(ctx, txHash)

	assert.Nil(t, result)
	assert.Error(t, err)
	client.AssertExpectations(t)
}

func TestWaitForConfirmations_InsufficientConfirmations(t *testing.T) {
	client := &MockWeb3Client{}
	txHash := common.HexToHash("0x12345")
	receipt := &types.Receipt{
		BlockNumber: big.NewInt(100),
	}

	client.On("TransactionReceipt", mock.Anything, txHash).Return(receipt, nil).Maybe()
	client.On("BlockNumber", mock.Anything).Return(uint64(102), nil).Maybe()

	tw := NewTransactionWatcher(client, 5*time.Second, 6)
	ctx := context.Background()

	result, err := tw.WaitForConfirmations(ctx, txHash)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for confirmations")
	client.AssertExpectations(t)
}
