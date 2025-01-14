package pyweb3

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient(t *testing.T) {
	t.Run("NewHTTPClient", func(t *testing.T) {
		t.Run("invalid URL", func(t *testing.T) {
			client, err := NewHTTPClient("invalid-url", "test-agent")
			assert.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "invalid URL")
		})

		t.Run("invalid scheme", func(t *testing.T) {
			client, err := NewHTTPClient("http://example.com", "test-agent")
			assert.Error(t, err)
			assert.Nil(t, client)
			assert.Contains(t, err.Error(), "URL scheme must be https")
		})

		t.Run("valid URL", func(t *testing.T) {
			client, err := NewHTTPClient("https://example.com", "test-agent")
			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, "example.com", client.domain)
			assert.Equal(t, DefaultHTTPSPort, client.portNum)
		})

		t.Run("custom port", func(t *testing.T) {
			client, err := NewHTTPClient("https://example.com:8443", "test-agent")
			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, 8443, client.portNum)
		})
	})

	t.Run("SendMessage", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		client, err := NewHTTPClient(server.URL, "test-agent")
		assert.NoError(t, err)

		t.Run("successful send", func(t *testing.T) {
			err := client.SendMessage(`{"method": "test"}`)
			assert.NoError(t, err)
		})

		t.Run("connection closed", func(t *testing.T) {
			client.Close()
			err := client.SendMessage(`{"method": "test"}`)
			assert.Error(t, err)
		})
	})

	t.Run("GetMessages", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response\r\n\r\n"))
		}))
		defer server.Close()

		client, err := NewHTTPClient(server.URL, "test-agent")
		assert.NoError(t, err)

		t.Run("successful receive", func(t *testing.T) {
			messages, err := client.GetMessages()
			assert.NoError(t, err)
			assert.NotNil(t, messages)
		})

		t.Run("closed connection", func(t *testing.T) {
			client.Close()
			messages, err := client.GetMessages()
			assert.NoError(t, err)
			assert.Nil(t, messages)
		})
	})
}
