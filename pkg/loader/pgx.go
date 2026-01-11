package loader

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/daschewie/foenixmgr/pkg/config"
	"github.com/daschewie/foenixmgr/pkg/protocol"
)

// PGXLoader loads PGX binary format files
// PGX format: "PGX" signature + version/CPU byte + 32-bit address + data
type PGXLoader struct {
	BaseLoader
	data   []byte
	config *config.Config
}

// NewPGXLoader creates a new PGX loader
func NewPGXLoader(cfg *config.Config) *PGXLoader {
	return &PGXLoader{
		config: cfg,
	}
}

// Open opens a PGX file
func (l *PGXLoader) Open(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	l.data = data
	return nil
}

// Close closes the PGX file (no-op for memory-loaded file)
func (l *PGXLoader) Close() error {
	l.data = nil
	return nil
}

// Process reads and parses the PGX file
func (l *PGXLoader) Process() error {
	if l.data == nil {
		return fmt.Errorf("file not open")
	}

	if l.handler == nil {
		return fmt.Errorf("handler not set")
	}

	// Check minimum size (signature + version + address = 8 bytes)
	if len(l.data) < protocol.PGXOffData {
		return fmt.Errorf("file too small to be valid PGX")
	}

	// Check signature
	signature := l.data[protocol.PGXOffSigStart:protocol.PGXOffSigEnd]
	if string(signature) != "PGX" {
		return fmt.Errorf("bad PGX signature: %s", signature)
	}

	// Check version
	versionByte := l.data[protocol.PGXOffVersion]
	pgxVersion := (versionByte >> 4) & 0x0F
	if pgxVersion > 0 {
		return fmt.Errorf("unsupported PGX version: %d", pgxVersion)
	}

	// Check CPU compatibility
	pgxCPU := versionByte & 0x0F
	if err := l.verifyCPUCompatibility(pgxCPU); err != nil {
		return err
	}

	// Get target address (32-bit little-endian)
	address := binary.LittleEndian.Uint32(l.data[protocol.PGXOffAddrStart:protocol.PGXOffAddrEnd])

	// Get data block
	block := l.data[protocol.PGXOffData:]

	// Send data to handler
	if err := l.handler(address, block); err != nil {
		return fmt.Errorf("failed to write data block: %w", err)
	}

	// Set up CPU-specific reset vectors
	if err := SetupResetVectors(l.config.CPU, address, l.handler); err != nil {
		return fmt.Errorf("failed to set up reset vectors: %w", err)
	}

	return nil
}

// verifyCPUCompatibility checks if the PGX file matches the configured CPU
func (l *PGXLoader) verifyCPUCompatibility(pgxCPU byte) error {
	switch pgxCPU {
	case protocol.PGXcpu65816:
		if l.config.CPU != "65816" {
			return fmt.Errorf("PGX is built for 65816, but CPU is configured as %s", l.config.CPU)
		}

	case protocol.PGXcpu65C02:
		if l.config.CPU != "65C02" && l.config.CPU != "65c02" {
			return fmt.Errorf("PGX is built for 65C02, but CPU is configured as %s", l.config.CPU)
		}

	case protocol.PGXcpu680X0:
		if !l.config.CPUIsMotorolatype680X0() {
			return fmt.Errorf("PGX is built for 680x0, but CPU is configured as %s", l.config.CPU)
		}

	default:
		return fmt.Errorf("unsupported PGX CPU type: 0x%02X", pgxCPU)
	}

	return nil
}
