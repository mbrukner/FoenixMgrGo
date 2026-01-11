package loader

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// SRecLoader loads Motorola SREC format files
type SRecLoader struct {
	BaseLoader
}

// NewSRecLoader creates a new SREC loader
func NewSRecLoader() *SRecLoader {
	return &SRecLoader{}
}

// Open opens a Motorola SREC file
func (l *SRecLoader) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	l.file = file
	return nil
}

// Process reads and parses the SREC file
// SREC format: S<type><count><address><data><checksum>
// Types: S0=header, S1=16-bit addr, S2=24-bit addr, S3=32-bit addr,
//        S7=32-bit start, S8=24-bit start, S9=16-bit start
func (l *SRecLoader) Process() error {
	if l.file == nil {
		return fmt.Errorf("file not open")
	}

	if l.handler == nil {
		return fmt.Errorf("handler not set")
	}

	// Regex pattern for SREC records
	pattern := regexp.MustCompile(`^S([0-9a-fA-F])([0-9a-fA-F]+)`)

	scanner := bufio.NewScanner(l.file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse the record type
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			return fmt.Errorf("invalid SREC format at line %d: %s", lineNum, line)
		}

		recordType, _ := strconv.ParseUint(matches[1], 16, 8)
		hexDigits := matches[2]

		switch recordType {
		case 0: // Header record - ignore
			continue

		case 1: // Data with 16-bit address
			if err := l.parseDataRecord(hexDigits, 2, lineNum); err != nil {
				return err
			}

		case 2: // Data with 24-bit address
			if err := l.parseDataRecord(hexDigits, 3, lineNum); err != nil {
				return err
			}

		case 3: // Data with 32-bit address
			if err := l.parseDataRecord(hexDigits, 4, lineNum); err != nil {
				return err
			}

		case 4: // Reserved
			continue

		case 5, 6: // Record count - ignore
			continue

		case 7, 8, 9: // Start address - ignore (not data)
			continue

		default:
			return fmt.Errorf("unsupported SREC type S%d at line %d", recordType, lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// parseDataRecord parses an SREC data record
// addressBytes: 2 for S1, 3 for S2, 4 for S3
func (l *SRecLoader) parseDataRecord(hexDigits string, addressBytes int, lineNum int) error {
	// Format: <count><address><data><checksum>
	// Count is 1 byte (2 hex digits)
	// Address is addressBytes (addressBytes*2 hex digits)
	// Data is remainder minus checksum (2 hex digits)

	if len(hexDigits) < 2+addressBytes*2+2 {
		return fmt.Errorf("SREC record too short at line %d", lineNum)
	}

	// Parse count (includes address, data, and checksum bytes)
	// count is not used in this implementation but is part of the SREC format
	_, _ = strconv.ParseUint(hexDigits[0:2], 16, 8)

	// Parse address
	addressHex := hexDigits[2 : 2+addressBytes*2]
	address, _ := strconv.ParseUint(addressHex, 16, 32)

	// Extract data (everything except count, address, and checksum)
	dataStart := 2 + addressBytes*2
	dataEnd := len(hexDigits) - 2 // Exclude checksum
	dataHex := hexDigits[dataStart:dataEnd]

	// Convert hex data to bytes
	data, err := hexStringToBytes(dataHex)
	if err != nil {
		return fmt.Errorf("invalid data at line %d: %w", lineNum, err)
	}

	// Send to handler
	if err := l.handler(uint32(address), data); err != nil {
		return fmt.Errorf("handler failed at line %d: %w", lineNum, err)
	}

	return nil
}
