// Package pyweb3 provides TLS socket functionality
// Copyright (C) 2021-2022 BitLogiK
package pyweb3

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"
)

const (
	// ReceivingBufferSize defines the size of the receiving buffer
	ReceivingBufferSize = 8192
	// DefaultTimeout is the default timeout for socket operations
	DefaultTimeout = 8 * time.Second
)

// TLSSocket represents a TLS socket client with a host, push and read data
type TLSSocket struct {
	conn *tls.Conn
}

// NewTLSSocket creates a new TLS connection with a host domain:port
func NewTLSSocket(domain string, port int) (*TLSSocket, error) {
	conf := &tls.Config{
		ServerName: domain,
	}

	addr := fmt.Sprintf("%s:%d", domain, port)
	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	log.Printf("Socket connected to %s", addr)

	// Set timeout
	err = conn.SetDeadline(time.Now().Add(DefaultTimeout))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set timeout: %w", err)
	}

	return &TLSSocket{
		conn: conn,
	}, nil
}

// Close closes the socket connection
func (t *TLSSocket) Close() error {
	if t.conn != nil {
		log.Println("Closing socket")
		err := t.conn.Close()
		t.conn = nil
		return err
	}
	return nil
}

// Send sends data to the host
func (t *TLSSocket) Send(data []byte) error {
	if t.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	_, err := t.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	return nil
}

// Receive reads data from the host.
// It's a blocking reception.
// If no data received after timeout: returns error
func (t *TLSSocket) Receive() ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("connection is closed")
	}

	buffer := make([]byte, ReceivingBufferSize)
	n, err := t.conn.Read(buffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Socket disconnected")
			t.Close()
			return nil, fmt.Errorf("connection closed by peer")
		}
		return nil, fmt.Errorf("failed to receive data: %w", err)
	}

	if n == 0 {
		log.Println("Socket disconnected")
		t.Close()
		return nil, fmt.Errorf("connection closed by peer")
	}

	return buffer[:n], nil
}
