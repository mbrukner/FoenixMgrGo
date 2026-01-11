package util

import "testing"

func TestCalculateCRC32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			expected: 0x00000000,
		},
		{
			name:     "Single byte",
			data:     []byte{0x00},
			expected: 0x00000000,
		},
		{
			name:     "Simple data",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			expected: 0x977824D1,
		},
		{
			name:     "Text data",
			data:     []byte("COPYFILE"),
			expected: 0xB792616B,
		},
		{
			name:     "Binary data",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: 0xDEBB20E3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCRC32(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateCRC32() = 0x%08X, want 0x%08X", result, tt.expected)
			}
		})
	}
}
