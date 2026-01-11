package loader

import (
	"fmt"
	"os"
)

// WDCLoader loads WDCTools binary format files
// Format: 'Z' signature followed by blocks of:
//   3-byte address (little-endian)
//   3-byte length (little-endian)
//   length bytes of data
// Terminates when address == 0
type WDCLoader struct {
	BaseLoader
	data []byte
}

// NewWDCLoader creates a new WDC binary loader
func NewWDCLoader() *WDCLoader {
	return &WDCLoader{}
}

// Open opens a WDC binary file
func (l *WDCLoader) Open(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Verify signature
	if len(data) < 1 || data[0] != 'Z' {
		return fmt.Errorf("invalid WDC file: missing 'Z' signature")
	}

	l.data = data
	return nil
}

// Close closes the WDC file (no-op for memory-loaded file)
func (l *WDCLoader) Close() error {
	l.data = nil
	return nil
}

// Process reads and parses the WDC binary file
func (l *WDCLoader) Process() error {
	if l.data == nil {
		return fmt.Errorf("file not open")
	}

	if l.handler == nil {
		return fmt.Errorf("handler not set")
	}

	offset := 1 // Skip 'Z' signature

	for offset < len(l.data) {
		address, block, newOffset, err := l.readBlock(offset)
		if err != nil {
			return err
		}

		offset = newOffset

		// Terminate if address is 0
		if address == 0 {
			break
		}

		// Send block to handler
		if err := l.handler(address, block); err != nil {
			return fmt.Errorf("handler failed: %w", err)
		}
	}

	return nil
}

// readBlock reads a single block from the WDC file
func (l *WDCLoader) readBlock(offset int) (uint32, []byte, int, error) {
	// Need at least 6 bytes for address (3) + length (3)
	if offset+6 > len(l.data) {
		return 0, nil, offset, fmt.Errorf("unexpected end of file at offset %d", offset)
	}

	// Read 3-byte address (little-endian)
	address := uint32(l.data[offset]) |
		uint32(l.data[offset+1])<<8 |
		uint32(l.data[offset+2])<<16

	// Read 3-byte length (little-endian)
	length := uint32(l.data[offset+3]) |
		uint32(l.data[offset+4])<<8 |
		uint32(l.data[offset+5])<<16

	offset += 6

	// If address is 0, this is the terminator block
	if address == 0 {
		return 0, nil, offset, nil
	}

	// Read data block
	if offset+int(length) > len(l.data) {
		return 0, nil, offset, fmt.Errorf("data block exceeds file size at offset %d", offset)
	}

	block := make([]byte, length)
	copy(block, l.data[offset:offset+int(length)])

	offset += int(length)

	return address, block, offset, nil
}
