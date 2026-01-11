package protocol

import "testing"

func TestCalculateLRC(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			name:     "Empty data",
			data:     []byte{},
			expected: 0x00,
		},
		{
			name:     "Single byte",
			data:     []byte{0x55},
			expected: 0x55,
		},
		{
			name:     "Two identical bytes",
			data:     []byte{0xAA, 0xAA},
			expected: 0x00, // XOR of same values = 0
		},
		{
			name:     "Header example",
			data:     []byte{0x55, 0x00, 0x00, 0x10, 0x00, 0x00, 0x10},
			expected: 0x55, // 0x55 ^ 0x00 ^ 0x00 ^ 0x10 ^ 0x00 ^ 0x00 ^ 0x10 = 0x55
		},
		{
			name:     "All ones",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: 0x00, // Even number of 0xFF = 0
		},
		{
			name:     "Mixed data",
			data:     []byte{0x12, 0x34, 0x56, 0x78},
			expected: 0x12 ^ 0x34 ^ 0x56 ^ 0x78,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLRC(tt.data)
			if result != tt.expected {
				t.Errorf("calculateLRC() = 0x%02X, want 0x%02X", result, tt.expected)
			}
		})
	}
}

func TestVerifyLRC(t *testing.T) {
	tests := []struct {
		name     string
		dataWithLRC []byte
		expected bool
	}{
		{
			name:     "Valid LRC",
			dataWithLRC: []byte{0x55, 0x00, 0x00, 0x10, 0x00, 0x00, 0x10, 0x55},
			expected: true,
		},
		{
			name:     "Invalid LRC",
			dataWithLRC: []byte{0x55, 0x00, 0x00, 0x10, 0x00, 0x00, 0x10, 0x00},
			expected: false,
		},
		{
			name:     "Empty data",
			dataWithLRC: []byte{},
			expected: false,
		},
		{
			name:     "Single byte",
			dataWithLRC: []byte{0x00},
			expected: false,
		},
		{
			name:     "Two bytes valid",
			dataWithLRC: []byte{0xAA, 0xAA},
			expected: true, // 0xAA ^ 0xAA = 0x00, but wait - this means data[0]=0xAA and LRC=0xAA, so calculateLRC([0xAA]) = 0xAA, expected 0xAA - valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyLRC(tt.dataWithLRC)
			if result != tt.expected {
				t.Errorf("verifyLRC() = %v, want %v", result, tt.expected)
			}
		})
	}
}
