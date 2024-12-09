package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type JSONRPCException struct {
	ErrorField interface{} `json:"error"`
}

func (e JSONRPCException) Error() string {
	return fmt.Sprintf("JSON-RPC Error: %v", e.ErrorField)
}

const (
	DEFAULT_USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:96.0) Gecko/20100101 Firefox/96.0"
)

// jsonEncode compacts JSON encoding.
func jsonEncode(dataObj interface{}) (string, error) {
	jsonData, err := json.Marshal(dataObj)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// jsonRPCUnpack decodes a JSON-RPC response and returns id and result.
func jsonRPCUnpack(buffer []byte) (string, interface{}, error) {
	var respObj map[string]interface{}
	err := json.Unmarshal(buffer, &respObj)
	if err != nil {
		return "", nil, fmt.Errorf("error: not JSON response: %s", string(buffer))
	}

	if respObj["jsonrpc"] != "2.0" {
		return "", nil, fmt.Errorf("server is not JSONRPC 2.0 but %v", respObj["jsonrpc"])
	}

	if errorField, ok := respObj["error"]; ok {
		return "", nil, JSONRPCException{ErrorField: errorField}
	}

	id, idOk := respObj["id"].(string)
	result, resultOk := respObj["result"]
	if !idOk || !resultOk {
		return "", nil, errors.New("response missing required fields 'id' or 'result'")
	}

	return id, result, nil
}

func main() {
	// Example usage
	log.Println("Starting JSON-RPC example...")

	exampleBuffer := []byte(`{"jsonrpc":"2.0","id":"1","result":"example result"}`)

	id, result, err := jsonRPCUnpack(exampleBuffer)
	if err != nil {
		log.Fatalf("Error unpacking JSON-RPC response: %v", err)
	}

	log.Printf("Response ID: %s, Result: %v", id, result)
}

