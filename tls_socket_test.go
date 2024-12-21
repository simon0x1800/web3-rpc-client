package pyweb3

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) (string, int, func()) {
	// Generate test certificate
	cert, err := tls.LoadX509KeyPair("testdata/server.crt", "testdata/server.key")
	if err != nil {
		t.Fatal(err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Start test server
	listener, err := tls.Listen("tcp", "127.0.0.1:0", config)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buffer := make([]byte, ReceivingBufferSize)
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		// Echo back the received data
		conn.Write(buffer[:n])
	}()

	addr := listener.Addr().(*net.TCPAddr)

	return "127.0.0.1", addr.Port, func() {
		listener.Close()
	}
}

func TestTLSSocket(t *testing.T) {
	domain, port, cleanup := setupTestServer(t)
	defer cleanup()

	// Create new TLS socket
	socket, err := NewTLSSocket(domain, port)
	assert.NoError(t, err)
	assert.NotNil(t, socket)
	defer socket.Close()

	// Test sending data
	testData := []byte("Hello, World!")
	err = socket.Send(testData)
	assert.NoError(t, err)

	// Test receiving data
	received, err := socket.Receive()
	assert.NoError(t, err)
	assert.Equal(t, testData, received)
}

func TestTLSSocketTimeout(t *testing.T) {
	domain, port, cleanup := setupTestServer(t)
	defer cleanup()

	socket, err := NewTLSSocket(domain, port)
	assert.NoError(t, err)
	defer socket.Close()

	// Test timeout
	time.Sleep(DefaultTimeout + time.Second)
	_, err = socket.Receive()
	assert.Error(t, err)
}
