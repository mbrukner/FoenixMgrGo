package loader

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// IntelHexLoader loads Intel HEX format files
type IntelHexLoader struct {
	BaseLoader
	baseAddress uint32
}

// NewIntelHexLoader creates a new Intel HEX loader
func NewIntelHexLoader() *IntelHexLoader {
	return &IntelHexLoader{
		baseAddress: 0,
	}
}

// Open opens an Intel HEX file
func (l *IntelHexLoader) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	l.file = file
	l.baseAddress = 0
	return nil
}

// Process reads and parses the Intel HEX file
// Intel HEX format: :LLAAAATT[DD...]CC
// LL = byte count, AAAA = address, TT = record type, DD = data, CC = checksum
func (l *IntelHexLoader) Process() error {
	if l.file == nil {
		return fmt.Errorf("file not open")
	}

	if l.handler == nil {
		return fmt.Errorf("handler not set")
	}

	// Regex pattern for Intel HEX records
	// :LLAAAATTDDDDCC where each letter is a hex digit
	pattern := regexp.MustCompile(`^:([0-9a-fA-F]{2})([0-9a-fA-F]{4})([0-9a-fA-F]{2})([0-9a-fA-F]*)([0-9a-fA-F]{2})`)

	scanner := bufio.NewScanner(l.file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse the line
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			return fmt.Errorf("invalid Intel HEX format at line %d: %s", lineNum, line)
		}

		// Extract fields
		byteCount, _ := strconv.ParseUint(matches[1], 16, 8)
		address, _ := strconv.ParseUint(matches[2], 16, 16)
		recordType, _ := strconv.ParseUint(matches[3], 16, 8)
		dataHex := matches[4]
		// checksum := matches[5] // Not verified in original Python

		switch recordType {
		case 0x00: // Data record
			// Convert hex string to bytes
			data, err := hexStringToBytes(dataHex)
			if err != nil {
				return fmt.Errorf("invalid data at line %d: %w", lineNum, err)
			}

			// Verify byte count
			if uint64(len(data)) != byteCount {
				return fmt.Errorf("byte count mismatch at line %d: expected %d, got %d",
					lineNum, byteCount, len(data))
			}

			// Send to handler with base address applied
			fullAddress := l.baseAddress + uint32(address)
			if err := l.handler(fullAddress, data); err != nil {
				return fmt.Errorf("handler failed at line %d: %w", lineNum, err)
			}

		case 0x01: // End of file
			// Terminate processing
			return nil

		case 0x02: // Extended segment address (80x86 real mode)
			// Set bits 4-19 of the base address
			// "Shitty 80x86 real mode addressing: take the address and do *16"
			segmentAddr, _ := strconv.ParseUint(dataHex, 16, 32)
			l.baseAddress = uint32(segmentAddr) << 4

		case 0x04: // Extended linear address
			// Set bits 16-31 of the base address
			extAddr, _ := strconv.ParseUint(dataHex, 16, 32)
			l.baseAddress = uint32(extAddr) << 16

		case 0x03, 0x05: // Start segment/linear address
			// Ignored (these specify execution start address, not data)

		default:
			return fmt.Errorf("unsupported record type 0x%02X at line %d", recordType, lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}
