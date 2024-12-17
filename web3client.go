package web3client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
)

type JSONRPCClient struct {
	NodeURL    string
	UserAgent  string
	Retries    int
}

func NewJSONRPCClient(nodeURL, userAgent string, retries int) *JSONRPCClient {
	return &JSONRPCClient{NodeURL: nodeURL, UserAgent: userAgent, Retries: retries}
}

func (c *JSONRPCClient) Request(method string, params interface{}) (string, error) {
	return "", nil
}

type Web3Client struct {
	JSONRPC *JSONRPCClient
}

func NewWeb3Client(nodeURL, userAgent string, retries int) *Web3Client {
	return &Web3Client{
		JSONRPC: NewJSONRPCClient(nodeURL, userAgent, retries),
	}
}

func (w *Web3Client) GetBalance(address, state string) (*big.Int, error) {
	balRaw, err := w.JSONRPC.Request("eth_getBalance", []interface{}{address, state})
	if err != nil {
		return nil, err
	}

	if len(balRaw) >= 2 && balRaw[:2] == "0x" {
		balance := new(big.Int)
		balance.SetString(balRaw[2:], 16)
		return balance, nil
	}

	return big.NewInt(0), nil
}

func (w *Web3Client) Call(contract, commandCode, data, state string) (string, error) {
	dataHex := fmt.Sprintf("0x%s%s", commandCode, data)
	params := []interface{}{
		map[string]interface{}{
			"to":   contract,
			"data": dataHex,
		},
		state,
	}
	return w.JSONRPC.Request("eth_call", params)
}

func (w *Web3Client) PushTx(txHex string) (string, error) {
	return w.JSONRPC.Request("eth_sendRawTransaction", []interface{}{"0x" + txHex})
}

func (w *Web3Client) GetTxNum(address, state string) (int64, error) {
	txCountRaw, err := w.JSONRPC.Request("eth_getTransactionCount", []interface{}{address, state})
	if err != nil {
		return 0, err
	}

	if len(txCountRaw) >= 2 && txCountRaw[:2] == "0x" {
		txCount, _ := new(big.Int).SetString(txCountRaw[2:], 16)
		return txCount.Int64(), nil
	}

	return 0, fmt.Errorf("bad data when reading getTransactionCount")
}

func (w *Web3Client) GetGasPrice() (*big.Int, error) {
	gasPriceRaw, err := w.JSONRPC.Request("eth_gasPrice", nil)
	if err != nil {
		return nil, err
	}

	if len(gasPriceRaw) >= 2 && gasPriceRaw[:2] == "0x" {
		gasPrice := new(big.Int)
		gasPrice.SetString(gasPriceRaw[2:], 16)
		return gasPrice, nil
	}

	return nil, fmt.Errorf("bad data when reading gasPrice")
}

func (w *Web3Client) GetLogs(param interface{}) (string, error) {
	return w.JSONRPC.Request("eth_getLogs", []interface{}{param})
}

func (w *Web3Client) SetFilter(param interface{}) (string, error) {
	return w.JSONRPC.Request("eth_newFilter", []interface{}{param})
}

func (w *Web3Client) GetFilter(filterID string) (string, error) {
	return w.JSONRPC.Request("eth_getFilterLogs", []interface{}{filterID})
}
