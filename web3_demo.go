package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	rpcHost    = "https://matic-mainnet.chainstacklabs.com"
	token0Call = "0dfe1681"
	token1Call = "d21220a7"
	decimalsCall = "313ce567"
	getReserves = "0902f1ac"
	symbolCall  = "95d89b41"
	ammPair     = "0xc2755915a85c6f6c1c0f3a86ac8c058f11caa9c9"
)

func readUint256(data []byte, offset int) *big.Int {
	return new(big.Int).SetBytes(data[offset : offset+32])
}

func readString(data string) (string, error) {
	dataBin, err := hex.DecodeString(data[2:])
	if err != nil {
		return "", err
	}
	strOffset := int(readUint256(dataBin, 0).Int64())
	strLen := int(readUint256(dataBin, strOffset).Int64())
	strOffset += 32
	return string(dataBin[strOffset : strOffset+strLen]), nil
}

type Web3Client struct {
	client *rpc.Client
}

func NewWeb3Client(host string) (*Web3Client, error) {
	client, err := rpc.Dial(host)
	if err != nil {
		return nil, err
	}
	return &Web3Client{client: client}, nil
}

func (c *Web3Client) Call(contractAddress common.Address, data string) (string, error) {
	var result string
	err := c.client.Call(&result, "eth_call", map[string]string{
		"to":   contractAddress.Hex(),
		"data": "0x" + data,
	}, "latest")
	if err != nil {
		return "", err
	}
	return result, nil
}

