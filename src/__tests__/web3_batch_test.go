package pyweb3

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWeb3Client is a mock implementation of Web3Client
type MockWeb3Client struct {
	mock.Mock
}

func (m *MockWeb3Client) SendTransaction(from, to common.Address, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(from, to, amount)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWeb3Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *MockWeb3Client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := m.Called(ctx, q, ch)
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

// MockSubscription implements ethereum.Subscription interface
type MockSubscription struct {
	mock.Mock
}

func (m *MockSubscription) Unsubscribe() {
	m.Called()
}

func (m *MockSubscription) Err() <-chan error {
	args := m.Called()
	return args.Get(0).(chan error)
}

func TestBatchProcessor_BatchTransfer(t *testing.T) {
	mockClient := new(MockWeb3Client)
	bp := NewBatchProcessor(mockClient, 10, 2)

	from := common.HexToAddress("0x1234")
	transfers := map[common.Address]*big.Int{
		common.HexToAddress("0x5678"): big.NewInt(1000),
		common.HexToAddress("0x9abc"): big.NewInt(2000),
	}

	// Setup mock expectations
	tx1 := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)
	tx2 := types.NewTransaction(1, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)

	mockClient.On("SendTransaction", from, common.HexToAddress("0x5678"), big.NewInt(1000)).Return(tx1, nil)
	mockClient.On("SendTransaction", from, common.HexToAddress("0x9abc"), big.NewInt(2000)).Return(tx2, nil)

	// Execute batch transfer
	results := bp.BatchTransfer(from, transfers)

	// Verify results
	assert.Equal(t, 2, len(results))
	mockClient.AssertExpectations(t)
}

func TestEventFilter(t *testing.T) {
	mockClient := new(MockWeb3Client)
	ef := NewEventFilter(mockClient)

	// Test SetAddresses
	addresses := []common.Address{common.HexToAddress("0x1234")}
	ef.SetAddresses(addresses)
	assert.Equal(t, addresses, ef.query.Addresses)

	// Test SetTopics
	topics := [][]common.Hash{{common.HexToHash("0x5678")}}
	ef.SetTopics(topics)
	assert.Equal(t, topics, ef.query.Topics)

	// Test SetBlockRange
	fromBlock := big.NewInt(100)
	toBlock := big.NewInt(200)
	ef.SetBlockRange(fromBlock, toBlock)
	assert.Equal(t, fromBlock, ef.query.FromBlock)
	assert.Equal(t, toBlock, ef.query.ToBlock)

	// Test GetLogs
	ctx := context.Background()
	expectedLogs := []types.Log{{
		Address: common.HexToAddress("0x1234"),
		Topics:  []common.Hash{common.HexToHash("0x5678")},
	}}

	mockClient.On("FilterLogs", ctx, ef.query).Return(expectedLogs, nil)
	logs, err := ef.GetLogs(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)

	// Test SubscribeLogs
	mockSub := new(MockSubscription)
	errCh := make(chan error)
	mockSub.On("Err").Return(errCh)

	mockClient.On("SubscribeFilterLogs", ctx, ef.query, mock.AnyChan()).Return(mockSub, nil)
	logCh, sub, err := ef.SubscribeLogs(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, logCh)
	assert.NotNil(t, sub)

	mockClient.AssertExpectations(t)
	mockSub.AssertExpectations(t)
}

func TestBatchProcessor_ConcurrencyLimit(t *testing.T) {
	mockClient := new(MockWeb3Client)
	concurrent := 2
	bp := NewBatchProcessor(mockClient, 10, concurrent)

	from := common.HexToAddress("0x1234")
	transfers := make(map[common.Address]*big.Int)

	// Create 5 transfers
	for i := 0; i < 5; i++ {
		addr := common.HexToAddress(fmt.Sprintf("0x%d", i))
		transfers[addr] = big.NewInt(int64(i+1) * 1000)
		tx := types.NewTransaction(uint64(i), common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)
		mockClient.On("SendTransaction", from, addr, transfers[addr]).Return(tx, nil)
	}

	// Add timing checks to verify concurrency
	start := time.Now()
	results := bp.BatchTransfer(from, transfers)
	duration := time.Since(start)

	// Verify results
	assert.Equal(t, 5, len(results))
	assert.True(t, duration >= time.Duration(5/concurrent)*100*time.Millisecond) // Assuming each transfer takes ~100ms

	mockClient.AssertExpectations(t)
}
