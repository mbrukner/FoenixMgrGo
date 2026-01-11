package util

import (
	"testing"
)

func TestParseHexAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
		wantErr  bool
	}{
		{"Simple hex", "1234", 0x1234, false},
		{"With 0x prefix", "0x1234", 0x1234, false},
		{"With $ prefix", "$1234", 0x1234, false},
		{"Uppercase", "ABCD", 0xABCD, false},
		{"Lowercase", "abcd", 0xABCD, false},
		{"24-bit address", "123456", 0x123456, false},
		{"Zero", "0", 0, false},
		{"Max 24-bit", "FFFFFF", 0xFFFFFF, false},
		{"Invalid characters", "GHIJ", 0, true},
		{"Empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseHexAddress(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHexAddress(%s) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseHexAddress(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseHexAddress(%s) = 0x%X, want 0x%X", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestParseHexSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint16
		wantErr  bool
	}{
		{"Simple hex", "10", 0x10, false},
		{"With 0x prefix", "0x20", 0x20, false},
		{"With $ prefix", "$30", 0x30, false},
		{"Uppercase", "FF", 0xFF, false},
		{"Lowercase", "ff", 0xFF, false},
		{"Zero", "0", 0, false},
		{"Max 16-bit", "FFFF", 0xFFFF, false},
		{"Invalid characters", "XYZ", 0, true},
		{"Empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseHexSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHexSize(%s) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseHexSize(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseHexSize(%s) = 0x%X, want 0x%X", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestHexDump(t *testing.T) {
	// Basic test to ensure HexDump doesn't panic
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, // "Hello Wo"
		0x72, 0x6C, 0x64, 0x21, 0x00, 0xFF}                // "rld!"

	// This is mainly a smoke test - we're just checking it doesn't panic
	HexDump(data, 0x1000)
}
