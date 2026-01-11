// Package util provides utility functions for FoenixMgr
package util

import (
	"os"
)

const stopFileName = "f256.stp"

// IsStopped returns true if the CPU is in a stopped state
// This is indicated by the presence of the f256.stp file
func IsStopped() bool {
	_, err := os.Stat(stopFileName)
	return err == nil // File exists = CPU is stopped
}

// SetStopIndicator creates the stop indicator file
// This marks the CPU as being in a stopped state
func SetStopIndicator() error {
	f, err := os.Create(stopFileName)
	if err != nil {
		return err
	}
	return f.Close()
}

// ClearStopIndicator removes the stop indicator file
// This marks the CPU as no longer being in a stopped state
func ClearStopIndicator() error {
	if !IsStopped() {
		return nil // Already clear
	}
	return os.Remove(stopFileName)
}
