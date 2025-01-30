package pyweb3

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

t.Run("NewHTTPClient", func(t *testing.T) {
	t.Run("empty URL", func(t *testing.T) {
		client, err := NewHTTPClient("", "test-agent")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("invalid port", func(t *testing.T) {
		client, err := NewHTTPClient("https://example.com:abcd", "test-agent")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
})

t.Run("SendMessage", func(t *testing.T) {
	t.Run("empty message", func(t *testing.T) {
		err := client.SendMessage("")
		assert.Error(t, err)
	})

	t.Run("large payload", func(t *testing.T) {
		largeMessage := strings.Repeat("A", 100000) // 100KB message
		err := client.SendMessage(largeMessage)
		assert.NoError(t, err)
	})
})

t.Run("GetMessages", func(t *testing.T) {
	serverError := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer serverError.Close()

	t.Run("server error response", func(t *testing.T) {
		client, err := NewHTTPClient(serverError.URL, "test-agent")
		assert.NoError(t, err)

		messages, err := client.GetMessages()
		assert.Error(t, err)
		assert.Nil(t, messages)
		assert.Contains(t, err.Error(), "error in response code")
	})

	t.Run("no response received", func(t *testing.T) {
		client.Close()
		messages, err := client.GetMessages()
		assert.NoError(t, err)
		assert.Nil(t, messages)
	})
})

t.Run("Close", func(t *testing.T) {
	client := &HTTPClient{}
	client.Close()
	assert.Nil(t, client.conn)
})
