package web3client

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEthClient is a mock implementation of the ethereum client interface
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEthClient) GetBalance(address common.Address) (*big.Int, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) SendTransaction(from, to common.Address, value *big.Int) (string, error) {
	args := m.Called(from, to, value)
	return args.String(0), args.Error(1)
}

func (m *MockEthClient) GetGasPrice() (*big.Int, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func TestWeb3Client(t *testing.T) {
	t.Run("Connection", func(t *testing.T) {
		t.Run("successful connection", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			mockClient.On("Connect").Return(nil)

			err := web3Client.Connect()
			
			assert.NoError(t, err)
			mockClient.AssertExpectations(t)
		})

		t.Run("connection error", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			expectedErr := errors.New("connection failed")
			mockClient.On("Connect").Return(expectedErr)

			err := web3Client.Connect()
			
			assert.Error(t, err)
			assert.Equal(t, expectedErr, err)
			mockClient.AssertExpectations(t)
		})
	})

	t.Run("GetBalance", func(t *testing.T) {
		t.Run("successful balance retrieval", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			address := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
			expectedBalance := big.NewInt(1000000000000000000) // 1 ETH in wei
			
			mockClient.On("GetBalance", address).Return(expectedBalance, nil)

			balance, err := web3Client.GetBalance(address)
			
			assert.NoError(t, err)
			assert.Equal(t, expectedBalance, balance)
			mockClient.AssertExpectations(t)
		})

		t.Run("invalid address", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			invalidAddress := common.HexToAddress("0x0")
			expectedErr := errors.New("invalid address")
			
			mockClient.On("GetBalance", invalidAddress).Return(nil, expectedErr)

			balance, err := web3Client.GetBalance(invalidAddress)
			
			assert.Error(t, err)
			assert.Nil(t, balance)
			assert.Equal(t, expectedErr, err)
			mockClient.AssertExpectations(t)
		})
	})

	t.Run("SendTransaction", func(t *testing.T) {
		t.Run("successful transaction", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
			to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44f")
			value := big.NewInt(1000000000000000000) // 1 ETH in wei
			expectedTxHash := "0x123..."

			mockClient.On("SendTransaction", from, to, value).Return(expectedTxHash, nil)

			txHash, err := web3Client.SendTransaction(from, to, value)
			
			assert.NoError(t, err)
			assert.Equal(t, expectedTxHash, txHash)
			mockClient.AssertExpectations(t)
		})

		t.Run("transaction error", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
			to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44f")
			value := big.NewInt(1000000000000000000)
			expectedErr := errors.New("insufficient funds")

			mockClient.On("SendTransaction", from, to, value).Return("", expectedErr)

			txHash, err := web3Client.SendTransaction(from, to, value)
			
			assert.Error(t, err)
			assert.Empty(t, txHash)
			assert.Equal(t, expectedErr, err)
			mockClient.AssertExpectations(t)
		})
	})

	t.Run("GetGasPrice", func(t *testing.T) {
		t.Run("successful gas price retrieval", func(t *testing.T) {
			mockClient := new(MockEthClient)
			web3Client := NewWeb3Client("http://localhost:8545", mockClient)

			expectedGasPrice := big.NewInt(20000000000) // 20 Gwei
			
			mockClient.On("GetGasPrice").Return(expectedGasPrice, nil)

			gasPrice, err := web3Client.GetGasPrice()
			
			assert.NoError(t, err)
			assert.Equal(t, expectedGasPrice, gasPrice)
			mockClient.AssertExpectations(t)
		})
	})
} 