package pyweb3

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEthClient mocks the ethereum client
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockEthClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, msg)
	return uint64(args.Int(0)), args.Error(1)
}

func TestWeb3Client(t *testing.T) {
	t.Run("NewWeb3Client", func(t *testing.T) {
		t.Run("invalid URL", func(t *testing.T) {
			client, err := NewWeb3Client("invalid-url")
			assert.Error(t, err)
			assert.Nil(t, client)
		})

		t.Run("valid URL", func(t *testing.T) {
			client, err := NewWeb3Client("https://localhost:8545")
			assert.NoError(t, err)
			assert.NotNil(t, client)
			client.Close()
		})
	})

	t.Run("GetLatestBlock", func(t *testing.T) {
		mockClient := new(MockEthClient)
		client := &Web3Client{
			client: mockClient,
			ctx:    context.Background(),
		}

		t.Run("successful retrieval", func(t *testing.T) {
			expectedBlock := types.NewBlock(
				&types.Header{Number: big.NewInt(12345)},
				[]*types.Transaction{},
				[]*types.Header{},
				[]*types.Receipt{},
				nil,
			)

			mockClient.On("BlockByNumber", mock.Anything, nil).Return(expectedBlock, nil).Once()

			block, err := client.GetLatestBlock()
			assert.NoError(t, err)
			assert.Equal(t, expectedBlock.Number().Uint64(), block.Number().Uint64())
		})

		t.Run("error case", func(t *testing.T) {
			mockClient.On("BlockByNumber", mock.Anything, nil).Return(nil, assert.AnError).Once()

			block, err := client.GetLatestBlock()
			assert.Error(t, err)
			assert.Nil(t, block)
		})
	})

	t.Run("WaitForTransaction", func(t *testing.T) {
		mockClient := new(MockEthClient)
		client := &Web3Client{
			client: mockClient,
			ctx:    context.Background(),
		}

		txHash := common.HexToHash("0x123")
		timeout := 2 * time.Second

		t.Run("successful receipt", func(t *testing.T) {
			expectedReceipt := &types.Receipt{
				Status:      1,
				BlockNumber: big.NewInt(12345),
			}

			mockClient.On("TransactionReceipt", mock.Anything, txHash).Return(expectedReceipt, nil).Once()

			receipt, err := client.WaitForTransaction(txHash, timeout)
			assert.NoError(t, err)
			assert.Equal(t, expectedReceipt.Status, receipt.Status)
		})

		t.Run("timeout", func(t *testing.T) {
			mockClient.On("TransactionReceipt", mock.Anything, txHash).Return(nil, assert.AnError).Once()

			receipt, err := client.WaitForTransaction(txHash, timeout)
			assert.Error(t, err)
			assert.Nil(t, receipt)
		})
	})

	t.Run("EstimateGas", func(t *testing.T) {
		mockClient := new(MockEthClient)
		client := &Web3Client{
			client: mockClient,
			ctx:    context.Background(),
		}

		to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
		data := []byte("test data")

		t.Run("successful estimation", func(t *testing.T) {
			expectedGas := uint64(21000)
			mockClient.On("EstimateGas", mock.Anything, mock.Anything).Return(expectedGas, nil).Once()

			gas, err := client.EstimateGas(to, data)
			assert.NoError(t, err)
			assert.Equal(t, expectedGas, gas)
		})

		t.Run("estimation error", func(t *testing.T) {
			mockClient.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(0), assert.AnError).Once()

			gas, err := client.EstimateGas(to, data)
			assert.Error(t, err)
			assert.Equal(t, uint64(0), gas)
		})
	})
}
