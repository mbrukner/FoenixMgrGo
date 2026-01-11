package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLabelFile(t *testing.T) {
	// Create a temporary label file
	tmpDir := t.TempDir()
	labelFile := filepath.Join(tmpDir, "test.lbl")

	labelContent := `; Test label file
; This is a comment
my_var = $1234
pointer = $5678
vector = $ABCD

; Another comment
data_block = $10000
`

	if err := os.WriteFile(labelFile, []byte(labelContent), 0644); err != nil {
		t.Fatalf("Failed to create test label file: %v", err)
	}

	// Test loading label file
	lf := NewLabelFile()
	if err := lf.Load(labelFile); err != nil {
		t.Fatalf("Failed to load label file: %v", err)
	}

	// Test label count
	if lf.Count() != 4 {
		t.Errorf("Expected 4 labels, got %d", lf.Count())
	}

	// Test looking up labels
	tests := []struct {
		label    string
		expected string
		wantErr  bool
	}{
		{"my_var", "1234", false},
		{"pointer", "5678", false},
		{"vector", "ABCD", false},
		{"data_block", "10000", false},
		{"nonexistent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			addr, err := lf.Lookup(tt.label)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for label %s, got nil", tt.label)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for label %s: %v", tt.label, err)
				}
				if addr != tt.expected {
					t.Errorf("Lookup(%s) = %s, want %s", tt.label, addr, tt.expected)
				}
			}
		})
	}
}

func TestLabelFileEmpty(t *testing.T) {
	// Create an empty label file
	tmpDir := t.TempDir()
	labelFile := filepath.Join(tmpDir, "empty.lbl")

	if err := os.WriteFile(labelFile, []byte("; Only comments\n"), 0644); err != nil {
		t.Fatalf("Failed to create test label file: %v", err)
	}

	lf := NewLabelFile()
	err := lf.Load(labelFile)
	if err == nil {
		t.Error("Expected error for empty label file, got nil")
	}
}

func TestLabelFileNotFound(t *testing.T) {
	lf := NewLabelFile()
	err := lf.Load("/nonexistent/path/file.lbl")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}
