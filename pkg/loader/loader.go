// Package loader provides file format loaders for various binary formats
// used by Foenix retro computers (Intel HEX, SREC, WDC, PGX, PGZ)
package loader

import (
	"fmt"
	"os"
)

// WriteHandler is a callback function that receives parsed address/data pairs
// This is typically connected to protocol.DebugPort.WriteBlock()
type WriteHandler func(address uint32, data []byte) error

// Loader defines the interface for all file format loaders
type Loader interface {
	// Open opens the file for reading
	Open(filename string) error

	// Close closes the file
	Close() error

	// SetHandler sets the callback function to receive parsed data
	SetHandler(handler WriteHandler)

	// Process reads and parses the file, invoking the handler for each block
	Process() error
}

// BaseLoader provides common functionality for all loaders
type BaseLoader struct {
	file    *os.File
	handler WriteHandler
}

// SetHandler sets the write handler callback
func (b *BaseLoader) SetHandler(handler WriteHandler) {
	b.handler = handler
}

// Close closes the file
func (b *BaseLoader) Close() error {
	if b.file != nil {
		return b.file.Close()
	}
	return nil
}

// Helper function to convert hex string to bytes
func hexStringToBytes(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("hex string length must be even")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, fmt.Errorf("invalid hex at position %d: %w", i, err)
		}
		bytes[i/2] = b
	}
	return bytes, nil
}
