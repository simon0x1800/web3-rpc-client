package jsonrpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONRPCRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		params         interface{}
		expectedID     int64
		expectedError  bool
		expectedStatus int
	}{
		{
			name:           "valid request",
			method:         "test_method",
			params:         map[string]string{"key": "value"},
			expectedID:     1,
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty method",
			method:         "",
			params:         nil,
			expectedID:     1,
			expectedError:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest(tt.method, tt.params)
			assert.Equal(t, "2.0", req.Version)
			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, tt.params, req.Params)
			assert.NotZero(t, req.ID)
		})
	}
}

func TestJSONRPCResponse(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      *Error
		expected Response
	}{
		{
			name:   "success response",
			result: "test result",
			err:    nil,
			expected: Response{
				Version: "2.0",
				Result:  "test result",
				Error:   nil,
				ID:      1,
			},
		},
		{
			name:   "error response",
			result: nil,
			err:    &Error{Code: -32600, Message: "Invalid Request"},
			expected: Response{
				Version: "2.0",
				Result:  nil,
				Error:   &Error{Code: -32600, Message: "Invalid Request"},
				ID:      1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewResponse(1, tt.result, tt.err)
			assert.Equal(t, tt.expected.Version, resp.Version)
			assert.Equal(t, tt.expected.Result, resp.Result)
			assert.Equal(t, tt.expected.Error, resp.Error)
			assert.Equal(t, tt.expected.ID, resp.ID)
		})
	}
}

func TestJSONRPCHandler(t *testing.T) {
	handler := func(params interface{}) (interface{}, *Error) {
		return "success", nil
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleFunc(w, r, handler)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		request        Request
		expectedStatus int
		expectedResult string
	}{
		{
			name: "valid request",
			request: Request{
				Version: "2.0",
				Method:  "test_method",
				Params:  map[string]string{"key": "value"},
				ID:      1,
			},
			expectedStatus: http.StatusOK,
			expectedResult: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response Response
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, response.Result)
		})
	}
}
