package pyweb3

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
)

// DefaultHTTPSPort is the default port for HTTPS connections
const DefaultHTTPSPort = 443

// HTTPClientException represents an error from the HTTP client
type HTTPClientException struct {
	message string
}

func (e *HTTPClientException) Error() string {
	return e.message
}

// HTTPClient handles HTTPS connections with TLS
type HTTPClient struct {
	conn             *tls.Conn
	receivedMessages [][]byte
	partialMessages  [][]byte
	portNum         int
	domain          string
	endpoint        string
	userAgent       string
}

// NewHTTPClient creates a new HTTPS client for a given URL
func NewHTTPClient(httpURL string, ua string) (*HTTPClient, error) {
	parsedURL, err := url.Parse(httpURL)
	if err != nil {
		return nil, &HTTPClientException{message: "invalid URL"}
	}

	if parsedURL.Scheme != "https" {
		return nil, &HTTPClientException{message: "URL scheme must be https"}
	}

	port := parsedURL.Port()
	portNum := DefaultHTTPSPort
	if port != "" {
		portNum = parseInt(port)
	}

	endpoint := parsedURL.Path
	if endpoint == "" {
		endpoint = "/"
	}

	return &HTTPClient{
		conn:             nil,
		receivedMessages: make([][]byte, 0),
		partialMessages:  make([][]byte, 0),
		portNum:         portNum,
		domain:          parsedURL.Hostname(),
		endpoint:        endpoint,
		userAgent:       ua,
	}, nil
}

// Close terminates the TLS connection
func (c *HTTPClient) Close() {
	if c.conn != nil {
		log.Printf("Closing TLS")
		c.conn.Close()
		c.conn = nil
	}
}
