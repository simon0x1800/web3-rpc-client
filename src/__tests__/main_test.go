package pyweb3_test

import (
	"errors"
	"math/big"
	"testing"

	web3client "web3-rpc-client/src"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestClientInitialization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := web3client.NewMockWeb3Client(ctrl)

	_, err := web3client.NewWeb3Client("https://...")
	assert.NoError(t, err, "Client should initialize without error")
}

func TestBatchProcessorInitialization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := web3client.NewMockWeb3Client(ctrl)

	batchProcessor := web3client.NewBatchProcessor(mockClient, 10, 5)
	assert.NotNil(t, batchProcessor, "BatchProcessor should initialize correctly")
}
