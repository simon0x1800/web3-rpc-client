package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

func TestNewWebSocketClient(t *testing.T) {
	// Override timeout constants for faster tests
	originalCyclesTimeout := cyclesTimeout
	originalUnitWaitingTime := unitWaitingTime
	defer func() {
		cyclesTimeout = originalCyclesTimeout
		unitWaitingTime = originalUnitWaitingTime
	}()
	cyclesTimeout = 2
	unitWaitingTime = 100 * time.Millisecond

	t.Run("invalid URL", func(t *testing.T) {
		_, err := NewWebSocketClient("invalid-url", "test-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid URL")
	})

	t.Run("invalid scheme", func(t *testing.T) {
		_, err := NewWebSocketClient("ws://example.com", "test-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid scheme, expected 'wss'")
	})

	t.Run("successful connection with handshake", func(t *testing.T) {
		// Create test server
		server := httptest.NewTLSServer(websocket.Handler(func(ws *websocket.Conn) {
			// Send "established" message
			websocket.Message.Send(ws, "established")
			// Keep connection alive for test duration
			select {}
		}))
		defer server.Close()

		// Convert http URL to wss
		wsURL := "wss" + server.URL[4:]

		client, err := NewWebSocketClient(wsURL, "test-agent")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Conn)

		client.Conn.Close()
	})

	t.Run("handshake timeout", func(t *testing.T) {
		server := httptest.NewTLSServer(websocket.Handler(func(ws *websocket.Conn) {
			// Don't send any message, let it timeout
			select {}
		}))
		defer server.Close()

		wsURL := "wss" + server.URL[4:]

		_, err := NewWebSocketClient(wsURL, "test-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WebSocket handshake timeout")
	})

	t.Run("handshake rejected", func(t *testing.T) {
		server := httptest.NewTLSServer(websocket.Handler(func(ws *websocket.Conn) {
			websocket.Message.Send(ws, "rejected")
			select {}
		}))
		defer server.Close()

		wsURL := "wss" + server.URL[4:]

		_, err := NewWebSocketClient(wsURL, "test-agent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WebSocket handshake rejected")
	})
}

func TestWebSocketClientException(t *testing.T) {
	t.Run("error with underlying error", func(t *testing.T) {
		underlyingErr := assert.AnError
		exc := &WebSocketClientException{
			Message: "test error",
			Err:     underlyingErr,
		}
		assert.Contains(t, exc.Error(), "test error")
		assert.Contains(t, exc.Error(), underlyingErr.Error())
	})

	t.Run("error without underlying error", func(t *testing.T) {
		exc := &WebSocketClientException{
			Message: "test error",
			Err:     nil,
		}
		assert.Equal(t, "test error", exc.Error())
	})
}

func TestWebSocketClientFields(t *testing.T) {
	client := &WebSocketClient{
		PartialTxtMsgs:   []string{"test1", "test2"},
		PartialBinMsgs:   [][]byte{[]byte("test1"), []byte("test2")},
		ReceivedMessages: []string{"msg1", "msg2"},
	}

	assert.Len(t, client.PartialTxtMsgs, 2)
	assert.Len(t, client.PartialBinMsgs, 2)
	assert.Len(t, client.ReceivedMessages, 2)
}

func TestWaitForHandshake(t *testing.T) {
	t.Run("successful handshake", func(t *testing.T) {
		client := &WebSocketClient{
			ReceivedMessages: []string{"established"},
		}
		err := client.waitForHandshake()
		assert.NoError(t, err)
	})

	t.Run("rejected handshake", func(t *testing.T) {
		client := &WebSocketClient{
			ReceivedMessages: []string{"rejected"},
		}
		err := client.waitForHandshake()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WebSocket handshake rejected")
	})

	t.Run("multiple messages before established", func(t *testing.T) {
		client := &WebSocketClient{
			ReceivedMessages: []string{"msg1", "msg2", "established"},
		}
		err := client.waitForHandshake()
		assert.NoError(t, err)
	})
}
