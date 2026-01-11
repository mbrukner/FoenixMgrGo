package loader

import (
	"fmt"
	"os"

	"github.com/daschewie/foenixmgr/pkg/config"
)

// PGZLoader loads PGZ (compressed) binary format files
// PGZ format: header byte ('z'=4-byte addr/size, 'Z'=3-byte) + blocks
// Each block: address + size + data
// Special: size=0 with addr>0 indicates start address
type PGZLoader struct {
	BaseLoader
	data        []byte
	config      *config.Config
	addressSize int
}

// NewPGZLoader creates a new PGZ loader
func NewPGZLoader(cfg *config.Config) *PGZLoader {
	return &PGZLoader{
		config: cfg,
	}
}

// Open opens a PGZ file
func (l *PGZLoader) Open(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check minimum size
	if len(data) < 1 {
		return fmt.Errorf("file too small to be valid PGZ")
	}

	// Determine address size from header byte
	switch data[0] {
	case 0x7A: // 'z' - 4-byte address and size fields
		l.addressSize = 4
	case 0x5A: // 'Z' - 3-byte address and size fields
		l.addressSize = 3
	default:
		return fmt.Errorf("invalid PGZ header: 0x%02X (expected 0x7A or 0x5A)", data[0])
	}

	l.data = data
	return nil
}

// Close closes the PGZ file (no-op for memory-loaded file)
func (l *PGZLoader) Close() error {
	l.data = nil
	return nil
}

// Process reads and parses the PGZ file
func (l *PGZLoader) Process() error {
	if l.data == nil {
		return fmt.Errorf("file not open")
	}

	if l.handler == nil {
		return fmt.Errorf("handler not set")
	}

	offset := 1 // Skip header byte

	for offset < len(l.data) {
		address, block, newOffset, err := l.readBlock(offset)
		if err != nil {
			return err
		}

		offset = newOffset

		// Check for terminator (address == 0)
		if address == 0 {
			break
		}

		// Check for start address block (size == 0, address > 0)
		if len(block) == 0 && address > 0 {
			// Set up CPU-specific reset vectors
			if err := SetupResetVectors(l.config.CPU, address, l.handler); err != nil {
				return fmt.Errorf("failed to set up reset vectors: %w", err)
			}
			continue
		}

		// Regular data block
		if address > 0 {
			// Support for large blocks: chunk into 1KB pieces
			const chunkSize = 1024
			if len(block) > chunkSize {
				totalLength := len(block)
				chunkOffset := 0

				for totalLength > 0 {
					currentChunkSize := chunkSize
					if totalLength < chunkSize {
						currentChunkSize = totalLength
					}

					chunk := block[chunkOffset : chunkOffset+currentChunkSize]
					if err := l.handler(address+uint32(chunkOffset), chunk); err != nil {
						return fmt.Errorf("failed to write chunk at 0x%X: %w", address+uint32(chunkOffset), err)
					}

					totalLength -= currentChunkSize
					chunkOffset += currentChunkSize
				}
			} else {
				// Small block, send directly
				if err := l.handler(address, block); err != nil {
					return fmt.Errorf("failed to write block at 0x%X: %w", address, err)
				}
			}
		}
	}

	return nil
}

// readBlock reads a single block from the PGZ file
// Returns: address, data block, new offset, error
func (l *PGZLoader) readBlock(offset int) (uint32, []byte, int, error) {
	// Need address + size fields (addressSize bytes each)
	requiredBytes := l.addressSize * 2
	if offset+requiredBytes > len(l.data) {
		return 0, nil, offset, fmt.Errorf("unexpected end of file at offset %d", offset)
	}

	// Read address (little-endian)
	address := l.readLittleEndianInt(l.data[offset:], l.addressSize)
	offset += l.addressSize

	// Read size (little-endian)
	size := l.readLittleEndianInt(l.data[offset:], l.addressSize)
	offset += l.addressSize

	// If address is 0, this is the terminator
	if address == 0 {
		return 0, nil, offset, nil
	}

	// If size is 0, this is a start address block (no data)
	if size == 0 {
		return address, nil, offset, nil
	}

	// Read data block
	if offset+int(size) > len(l.data) {
		return 0, nil, offset, fmt.Errorf("data block exceeds file size at offset %d", offset)
	}

	block := make([]byte, size)
	copy(block, l.data[offset:offset+int(size)])

	offset += int(size)

	return address, block, offset, nil
}

// readLittleEndianInt reads a little-endian integer of the specified byte count
func (l *PGZLoader) readLittleEndianInt(data []byte, byteCount int) uint32 {
	var value uint32
	for i := 0; i < byteCount && i < len(data); i++ {
		value |= uint32(data[i]) << (i * 8)
	}
	return value
}
