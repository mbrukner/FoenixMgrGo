package util

import (
	"fmt"
	"os"
	"strings"
)

// HexDump displays a block of memory in hex dump format
// Shows address, hex bytes, and ASCII representation
func HexDump(data []byte, startAddress uint32) {
	const bytesPerLine = 16

	for offset := 0; offset < len(data); offset += bytesPerLine {
		// Calculate address for this line
		address := startAddress + uint32(offset)

		// Print address
		fmt.Printf("%06X: ", address)

		// Print hex bytes
		lineEnd := offset + bytesPerLine
		if lineEnd > len(data) {
			lineEnd = len(data)
		}

		for i := offset; i < lineEnd; i++ {
			fmt.Printf("%02X ", data[i])
		}

		// Pad with spaces if this is the last line
		for i := lineEnd; i < offset+bytesPerLine; i++ {
			fmt.Print("   ")
		}

		// Print ASCII representation
		fmt.Print(" | ")
		for i := offset; i < lineEnd; i++ {
			b := data[i]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}

		fmt.Println()
	}
}

// FormatHex formats a byte slice as a hex string
func FormatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, b := range data {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%02X", b))
	}
	return sb.String()
}

// ParseHexAddress parses a hexadecimal address string (with or without 0x/$ prefix)
func ParseHexAddress(s string) (uint32, error) {
	// Remove 0x or $ prefix if present
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimPrefix(s, "$")

	var addr uint32
	_, err := fmt.Sscanf(s, "%x", &addr)
	if err != nil {
		return 0, fmt.Errorf("invalid hex address '%s': %w", s, err)
	}
	return addr, nil
}

// ParseHexSize parses a hexadecimal size string (with or without 0x/$ prefix)
func ParseHexSize(s string) (uint16, error) {
	// Remove 0x or $ prefix if present
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimPrefix(s, "$")

	var size uint16
	_, err := fmt.Sscanf(s, "%x", &size)
	if err != nil {
		return 0, fmt.Errorf("invalid hex size '%s': %w", s, err)
	}
	return size, nil
}

// ReadFile reads an entire file and returns its contents
func ReadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return data, nil
}
