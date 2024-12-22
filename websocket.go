package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"golang.org/x/net/websocket"
)

const (
	defaultHTTPSPort  = 443
	globalTimeout     = 8 * time.Second
	unitWaitingTime   = 400 * time.Millisecond
)

var cyclesTimeout = int(globalTimeout / unitWaitingTime)

type WebSocketClientException struct {
	Message string
	Err     error
}

func (e *WebSocketClientException) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

type WebSocketClient struct {
	Conn            *websocket.Conn
	PartialTxtMsgs  []string
	PartialBinMsgs  [][]byte
	ReceivedMessages []string
}

func NewWebSocketClient(wsURL, userAgent string) (*WebSocketClient, error) {
	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		return nil, &WebSocketClientException{"Invalid URL", err}
	}

	if parsedURL.Scheme != "wss" {
		return nil, &WebSocketClientException{"Invalid scheme, expected 'wss'", nil}
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = fmt.Sprintf("%d", defaultHTTPSPort)
	}

	tlsConfig := &tls.Config{ServerName: host}
	origin := "https://" + host
	wsEndpoint := parsedURL.Path
	if parsedURL.RawQuery != "" {
		wsEndpoint += "?" + parsedURL.RawQuery
	}

	log.Printf("Connecting to WebSocket Host=%s PathTarget=%s", host, wsEndpoint)

	conn, err := websocket.DialConfig(&websocket.Config{
		Location: &url.URL{Scheme: "wss", Host: host + ":" + port, Path: wsEndpoint},
		Origin:   &url.URL{Scheme: "https", Host: host},
		TlsConfig: tlsConfig,
		Header: map[string][]string{
			"User-Agent": {userAgent},
		},
	})
	if err != nil {
		return nil, &WebSocketClientException{"Error during WebSocket connection", err}
	}

	client := &WebSocketClient{
		Conn:            conn,
		PartialTxtMsgs:  []string{},
		PartialBinMsgs:  [][]byte{},
		ReceivedMessages: []string{},
	}

	if err := client.waitForHandshake(); err != nil {
		return nil, err
	}

	return client, nil
}

func (client *WebSocketClient) waitForHandshake() error {
	for cycle := 0; cycle < cyclesTimeout; cycle++ {
		log.Printf("Waiting WebSocket handshake: %dth loop.", cycle+1)
		time.Sleep(unitWaitingTime)
		if len(client.ReceivedMessages) > 0 {
			for len(client.ReceivedMessages) > 0 {
				msg := client.ReceivedMessages[len(client.ReceivedMessages)-1]
				client.ReceivedMessages = client.ReceivedMessages[:len(client.ReceivedMessages)-1]

				if msg == "established" {
					return nil
				}
				if msg == "rejected" {
					return &WebSocketClientException{"WebSocket handshake rejected", nil}
				}
			}
		}
	}
	return &WebSocketClientException{"WebSocket handshake timeout", nil}
}

func main() {
	wsURL := "wss://example.com/socket"
	userAgent := "GoWebSocketClient/1.0"

	client, err := NewWebSocketClient(wsURL, userAgent)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer client.Conn.Close()
	log.Println("WebSocket connection established successfully.")
}
