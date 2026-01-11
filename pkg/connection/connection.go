// Package connection provides connection abstractions for communicating with Foenix hardware
package connection

import (
	"fmt"
	"strings"
)

// Connection defines the interface for communicating with Foenix debug port
// Implementations include serial port and TCP socket connections
type Connection interface {
	// Open establishes the connection
	Open(port string) error

	// Close terminates the connection
	Close() error

	// IsOpen returns true if the connection is currently open
	IsOpen() bool

	// Read reads exactly n bytes from the connection
	// Returns error if fewer bytes are available
	Read(n int) ([]byte, error)

	// Write writes all data to the connection
	// Returns number of bytes written and error
	Write(data []byte) (int, error)
}

// NewConnection creates the appropriate connection type based on the port string
// If port contains ':', creates a TCP connection (e.g., "192.168.1.114:2560")
// Otherwise, creates a serial port connection (e.g., "COM3", "/dev/ttyUSB0")
func NewConnection(port string) Connection {
	if strings.Contains(port, ":") {
		// TCP connection detected
		return &TCPConnection{}
	}
	// Serial connection
	return &SerialConnection{}
}

// ValidatePort performs basic validation on a port string
func ValidatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	return nil
}
