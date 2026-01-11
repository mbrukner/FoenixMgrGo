package util

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// LabelFile represents a 64TASS label file parser
type LabelFile struct {
	labels map[string]string // label name -> hex address (without $)
}

// NewLabelFile creates a new label file parser
func NewLabelFile() *LabelFile {
	return &LabelFile{
		labels: make(map[string]string),
	}
}

// Load parses a 64TASS label file
// Format: LABEL = $ADDRESS
func (lf *LabelFile) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open label file: %w", err)
	}
	defer file.Close()

	// Regex pattern for label lines: LABEL = $ADDRESS
	// Example: "my_var = $1234"
	pattern := regexp.MustCompile(`^(\S+)\s*=\s*\$(\S+)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Try to match label pattern
		matches := pattern.FindStringSubmatch(line)
		if matches != nil {
			label := matches[1]
			address := matches[2]
			lf.labels[label] = address
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading label file: %w", err)
	}

	if len(lf.labels) == 0 {
		return fmt.Errorf("no labels found in file")
	}

	return nil
}

// Lookup finds the address for a given label
func (lf *LabelFile) Lookup(label string) (string, error) {
	address, ok := lf.labels[label]
	if !ok {
		return "", fmt.Errorf("label '%s' not found in label file", label)
	}
	return address, nil
}

// Count returns the number of labels loaded
func (lf *LabelFile) Count() int {
	return len(lf.labels)
}
