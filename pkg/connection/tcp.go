package connection

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// TCPConnection implements Connection interface for TCP socket communication
// Used for connecting to a TCP-to-serial bridge
type TCPConnection struct {
	conn   net.Conn
	isOpen bool
}

// Open establishes a TCP connection to the specified host:port
func (t *TCPConnection) Open(port string) error {
	parts := strings.Split(port, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid TCP address format (expected host:port): %s", port)
	}

	host := parts[0]
	tcpPort := parts[1]

	address := net.JoinHostPort(host, tcpPort)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	t.conn = conn
	t.isOpen = true
	return nil
}

// Close closes the TCP connection
func (t *TCPConnection) Close() error {
	if t.conn == nil {
		return nil
	}
	t.isOpen = false
	return t.conn.Close()
}

// IsOpen returns true if the connection is currently open
func (t *TCPConnection) IsOpen() bool {
	return t.isOpen
}

// Read reads exactly n bytes from the TCP connection
func (t *TCPConnection) Read(n int) ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("TCP connection not open")
	}

	buf := make([]byte, n)
	totalRead := 0

	for totalRead < n {
		bytesRead, err := t.conn.Read(buf[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("TCP read error: %w", err)
		}
		if bytesRead == 0 {
			return nil, fmt.Errorf("TCP connection closed")
		}
		totalRead += bytesRead
	}

	return buf, nil
}

// Write writes all data to the TCP connection
func (t *TCPConnection) Write(data []byte) (int, error) {
	if t.conn == nil {
		return 0, fmt.Errorf("TCP connection not open")
	}

	totalWritten := 0
	for totalWritten < len(data) {
		n, err := t.conn.Write(data[totalWritten:])
		if err != nil {
			return totalWritten, fmt.Errorf("TCP write error: %w", err)
		}
		totalWritten += n
	}

	return totalWritten, nil
}
