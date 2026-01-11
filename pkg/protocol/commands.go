// Package protocol implements the Foenix debug port binary protocol
package protocol

import "time"

// Debug port protocol commands
const (
	// Memory operations
	CMDReadMem  = 0x00 // Read from memory
	CMDWriteMem = 0x01 // Write to memory

	// Flash operations
	CMDProgramFlash  = 0x10 // Program entire flash from RAM
	CMDEraseFlash    = 0x11 // Erase entire flash
	CMDEraseSector   = 0x12 // Erase 4KB sector
	CMDProgramSector = 0x13 // Program 4KB sector from RAM

	// CPU control (F256 only)
	CMDStopCPU  = 0x20 // Stop CPU execution
	CMDStartCPU = 0x21 // Start CPU execution

	// Debug mode
	CMDEnterDebug = 0x80 // Enter debug mode (stops CPU, enables commands)
	CMDExitDebug  = 0x81 // Exit debug mode (resets CPU)

	// Boot source (F256k)
	CMDBootRAM   = 0x90 // Boot from RAM LUTs
	CMDBootFlash = 0x91 // Boot from Flash LUTs

	// System information
	CMDRevision = 0xFE // Get debug interface revision
)

// Protocol sync bytes
const (
	RequestSyncByte  = 0x55 // Sent at start of each request
	ResponseSyncByte = 0xAA // Expected at start of each response
)

// Timing delays for flash operations
const (
	DelayEraseSector   = 1 * time.Second // Delay after ERASE_SECTOR command
	DelayProgramSector = 2 * time.Second // Delay after PROGRAM_SECTOR command
)

// Boot source identifiers (for F256jr Rev A)
const (
	BootSrcRAM   = 0x00 // Boot from RAM
	BootSrcFlash = 0x01 // Boot from Flash
)

// PGX file format constants
const (
	// CPU type identifiers in PGX header
	PGXcpu65816 = 0x01 // 65816 CPU
	PGXcpu680X0 = 0x02 // 680x0 family CPU
	PGXcpu65C02 = 0x03 // 65C02 CPU

	// PGX header offsets
	PGXOffSigStart  = 0 // Start of "PGX" signature
	PGXOffSigEnd    = 3 // End of signature
	PGXOffVersion   = 3 // Version nibble and CPU type nibble
	PGXOffAddrStart = 4 // Start of 32-bit address (little-endian)
	PGXOffAddrEnd   = 8 // End of address
	PGXOffData      = 8 // Start of program data
)
