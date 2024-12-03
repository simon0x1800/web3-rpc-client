package pyweb3

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"strings"
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

func (c *HttpClient) SendMessage(message string) error {
	// Establish TLS connection
	log.Printf("Connecting to HTTPS Host: %s Port: %d", c.domain, c.portNum)
	
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.domain, c.portNum), &tls.Config{})
	if err != nil {
		log.Printf("Error during TLS connection: %v", err)
		return fmt.Errorf("tls connection error: %w", err)
	}
	c.tlsConn = conn
	defer c.tlsConn.Close()

	log.Printf("Connected to HTTPS Host=%s PathTarget=%s", c.domain, c.endpoint)

	// Construct HTTP POST request
	headers := []string{
		fmt.Sprintf("Host: %s", c.domain),
		fmt.Sprintf("User-Agent: %s", c.userAgent),
		"Connection: close",
		"Content-Type: application/json",
		fmt.Sprintf("Content-Length: %d", len(message)),
	}

	request := fmt.Sprintf("POST %s HTTP/1.1\r\n%s\r\n\r\n%s",
		c.endpoint,
		strings.Join(headers, "\r\n"),
		message,
	)

	log.Printf("Sending HTTP POST data: %s", request)
	
	// Send the request
	_, err = c.tlsConn.Write([]byte(request))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	return nil
}
