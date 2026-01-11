package loader

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/config"
)

// SetupResetVectors configures CPU-specific reset vectors for the given start address
// This is called by PGX and PGZ loaders to enable the program to start on reset
func SetupResetVectors(cpu string, startAddress uint32, handler WriteHandler) error {
	switch cpu {
	case "65816":
		return setup65816Vectors(startAddress, handler)

	case "65c02", "65C02":
		return setup65C02Vectors(startAddress, handler)

	default:
		// Check if it's a 680x0 variant
		cfg := &config.Config{CPU: cpu}
		if cfg.CPUIsMotorolatype680X0() {
			return setup680X0Vectors(startAddress, handler)
		}
		return fmt.Errorf("unsupported CPU type: %s", cpu)
	}
}

// setup65816Vectors sets up reset vectors for the 65816 CPU
func setup65816Vectors(startAddress uint32, handler WriteHandler) error {
	if startAddress&0xFF0000 != 0 {
		// Startup code is not in bank 0, need a stub at 0xFF80:
		//   CLC       ; 0x18
		//   XCE       ; 0xFB
		//   JML addr  ; 0x5C <low> <mid> <high>
		stub := []byte{
			0x18, // CLC
			0xFB, // XCE
			0x5C, // JML
			byte(startAddress),
			byte(startAddress >> 8),
			byte(startAddress >> 16),
		}
		if err := handler(0xFF80, stub); err != nil {
			return fmt.Errorf("failed to write 65816 stub: %w", err)
		}

		// Point reset vector to stub
		resetVector := []byte{0x80, 0xFF}
		if err := handler(0xFFFC, resetVector); err != nil {
			return fmt.Errorf("failed to write 65816 reset vector: %w", err)
		}
	} else {
		// Startup code is in bank 0, point reset vector directly to it
		resetVector := []byte{
			byte(startAddress),
			byte(startAddress >> 8),
		}
		if err := handler(0xFFFC, resetVector); err != nil {
			return fmt.Errorf("failed to write 65816 reset vector: %w", err)
		}
	}

	return nil
}

// setup65C02Vectors sets up reset vectors for the 65C02 CPU
func setup65C02Vectors(startAddress uint32, handler WriteHandler) error {
	// Point reset vector to start address
	resetVector := []byte{
		byte(startAddress),
		byte(startAddress >> 8),
	}
	if err := handler(0xFFFC, resetVector); err != nil {
		return fmt.Errorf("failed to write 65C02 reset vector: %w", err)
	}

	// "CROSSDEV" springboard for microkernel support
	crossdev := []byte{0x43, 0x52, 0x4F, 0x53, 0x53, 0x44, 0x45, 0x56}
	if err := handler(0x0080, crossdev); err != nil {
		return fmt.Errorf("failed to write CROSSDEV signature: %w", err)
	}

	// Microkernel start address
	startAddr := []byte{
		byte(startAddress),
		byte(startAddress >> 8),
	}
	if err := handler(0x0088, startAddr); err != nil {
		return fmt.Errorf("failed to write microkernel start address: %w", err)
	}

	// Kernel args extlen (0 until argument passing is implemented)
	kernelArgs := []byte{0x00, 0x00}
	if err := handler(0x00FA, kernelArgs); err != nil {
		return fmt.Errorf("failed to write kernel args: %w", err)
	}

	return nil
}

// setup680X0Vectors sets up reset vectors for 680x0 CPUs (68000, 68040, 68060)
func setup680X0Vectors(startAddress uint32, handler WriteHandler) error {
	// Reset vector at address 0x00000004 (32-bit, big-endian)
	resetVector := []byte{
		byte(startAddress >> 24),
		byte(startAddress >> 16),
		byte(startAddress >> 8),
		byte(startAddress),
	}

	if err := handler(0x00000004, resetVector); err != nil {
		return fmt.Errorf("failed to write 680x0 reset vector: %w", err)
	}

	return nil
}
