// Package config provides configuration management for FoenixMgr.
// It reads settings from foenixmgr.ini using multiple search paths.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// Config holds all configuration settings for FoenixMgr
type Config struct {
	// Serial/connection settings
	Port     string
	DataRate int
	Timeout  int

	// Hardware settings
	CPU       string
	ChunkSize int
	FlashSize int

	// Development settings
	LabelFile string
	Address   string

	// Machine-specific settings (set via SetTarget)
	flashPageSize   int
	flashSectorSize int
	ramSize         int
}

// Load reads configuration from foenixmgr.ini in the following search order:
// 1. Current directory (./foenixmgr.ini)
// 2. $FOENIXMGR directory ($FOENIXMGR/foenixmgr.ini)
// 3. Home directory (~/foenixmgr.ini)
func Load() (*Config, error) {
	// Build list of paths to search
	var searchPaths []string

	// 1. Current directory
	searchPaths = append(searchPaths, filepath.Join(".", "foenixmgr.ini"))

	// 2. $FOENIXMGR directory
	if foenixDir := os.Getenv("FOENIXMGR"); foenixDir != "" {
		searchPaths = append(searchPaths, filepath.Join(foenixDir, "foenixmgr.ini"))
	}

	// 3. Home directory
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(home, "foenixmgr.ini"))
	}

	// Try each path
	var iniFile *ini.File
	var configPath string
	var err error

	for _, path := range searchPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			iniFile, err = ini.Load(path)
			if err == nil {
				configPath = path
				break
			}
		}
	}

	if iniFile == nil {
		return nil, fmt.Errorf("no foenixmgr.ini file found in current directory, $FOENIXMGR, or home directory")
	}

	// Get DEFAULT section
	section := iniFile.Section("DEFAULT")

	// Create config with defaults
	cfg := &Config{
		Port:      section.Key("port").MustString("COM3"),
		DataRate:  section.Key("data_rate").MustInt(6000000),
		Timeout:   section.Key("timeout").MustInt(60),
		CPU:       section.Key("cpu").MustString("65c02"),
		ChunkSize: section.Key("chunk_size").MustInt(4096),
		FlashSize: section.Key("flash_size").MustInt(524288),
		LabelFile: section.Key("labels").MustString("basic8"),
		Address:   section.Key("address").MustString("380000"),
	}

	_ = configPath // Used for debugging if needed

	return cfg, nil
}

// SetTarget configures machine-specific flash and RAM settings
func (c *Config) SetTarget(machineName string) {
	machineName = strings.ToLower(machineName)

	// Reset to defaults
	c.flashPageSize = 0
	c.flashSectorSize = 0
	c.ramSize = 8

	switch machineName {
	case "fnx1591":
		c.flashPageSize = 8
		c.ramSize = 8
		c.flashSectorSize = 32

	case "f256k", "f256jr":
		c.flashPageSize = 8
		c.ramSize = 8
		c.flashSectorSize = 8
	}
}

// CPUIsMotorolatype680X0 returns true if the CPU is any Motorola 680x0 variant
func (c *Config) CPUIsMotorolatype680X0() bool {
	cpu := c.CPU
	return cpu == "m68k" || cpu == "68000" || cpu == "68040" || cpu == "68060"
}

// CPUIsM68k32 returns true if the CPU is a 32-bit Motorola 680x0 (68040 or 68060)
// These CPUs require 4-byte aligned memory operations
func (c *Config) CPUIsM68k32() bool {
	cpu := c.CPU
	return cpu == "68040" || cpu == "68060"
}

// FlashPageSize returns the size of the largest block of memory that can be
// copied to flash at one time (in KB). Returns 0 if machine doesn't support
// paged programming.
func (c *Config) FlashPageSize() int {
	return c.flashPageSize
}

// FlashSectorSize returns the size of the flash sector (in KB).
// Returns 0 if machine doesn't support paged programming.
func (c *Config) FlashSectorSize() int {
	return c.flashSectorSize
}

// RAMSize returns the number of bytes in RAM that can be used to write to flash (in KB)
func (c *Config) RAMSize() int {
	return c.ramSize
}

// ConfigPath returns the path to the config file that was loaded
func ConfigPath() (string, error) {
	// Check each location in order
	paths := []string{
		filepath.Join(".", "foenixmgr.ini"),
	}

	if foenixDir := os.Getenv("FOENIXMGR"); foenixDir != "" {
		paths = append(paths, filepath.Join(foenixDir, "foenixmgr.ini"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "foenixmgr.ini"))
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no foenixmgr.ini file found")
}
