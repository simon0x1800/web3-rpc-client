package pyweb3

import (
	"log"
	"math/big"

	web3client "web3-rpc-client/src"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	// Initialize Web3 client
	client, err := web3client.NewWeb3Client("https://...")
	if err != nil {
		log.Fatal(err)
	}

	// Create a batch processor
	batchProcessor := web3client.NewBatchProcessor(client, 10, 5)

	// Example batch transfer
	from := common.HexToAddress("0x123...")
	transfers := map[common.Address]*big.Int{
		common.HexToAddress("0x456..."): big.NewInt(1e18),
		common.HexToAddress("0x789..."): big.NewInt(2e18),
	}

	results := batchProcessor.BatchTransfer(from, transfers)
	for _, result := range results {
		if result.Error != nil {
			log.Printf("Transfer to %s failed: %v", result.To.Hex(), result.Error)
		} else {
			log.Printf("Transfer to %s successful: %s", result.To.Hex(), result.TxHash.Hex())
		}
	}
}
