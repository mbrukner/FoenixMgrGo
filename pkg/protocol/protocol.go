package protocol

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/daschewie/foenixmgr/pkg/config"
	"github.com/daschewie/foenixmgr/pkg/connection"
)

// DebugPort provides the main interface for communicating with Foenix hardware
type DebugPort struct {
	conn    connection.Connection
	config  *config.Config
	status0 byte
	status1 byte
}

// NewDebugPort creates a new DebugPort instance
func NewDebugPort(conn connection.Connection, cfg *config.Config) *DebugPort {
	return &DebugPort{
		conn:   conn,
		config: cfg,
	}
}

// IsOpen returns true if the connection is currently open
func (dp *DebugPort) IsOpen() bool {
	return dp.conn.IsOpen()
}

// Close closes the connection to the Foenix hardware
func (dp *DebugPort) Close() error {
	return dp.conn.Close()
}

// GetStatus0 returns the first status byte from the last operation
func (dp *DebugPort) GetStatus0() byte {
	return dp.status0
}

// GetStatus1 returns the second status byte from the last operation
func (dp *DebugPort) GetStatus1() byte {
	return dp.status1
}

// transfer sends a command to the Foenix debug port and receives the response
// This is the core protocol method that handles the binary protocol communication
//
// Request packet format (7-byte header + data + 1-byte LRC):
//   [0x55][CMD][ADDR_HI][ADDR_MID][ADDR_LO][LEN_HI][LEN_LO][...DATA...][LRC]
//
// Response packet format:
//   [0xAA][STATUS0][STATUS1][...DATA...][LRC]
func (dp *DebugPort) transfer(command byte, address uint32, data []byte, readLength uint16) ([]byte, error) {
	// Reset status bytes
	dp.status0 = 0
	dp.status1 = 0

	// Determine length
	length := readLength
	if data != nil && len(data) > 0 {
		length = uint16(len(data))
	}

	// Build 7-byte header
	header := make([]byte, 7)
	header[0] = RequestSyncByte
	header[1] = command

	// Address is 24-bit (3 bytes), big-endian
	header[2] = byte(address >> 16)
	header[3] = byte(address >> 8)
	header[4] = byte(address)

	// Length is 16-bit (2 bytes), big-endian
	binary.BigEndian.PutUint16(header[5:7], length)

	// Calculate LRC checksum (XOR of bytes 0-6, excluding sync byte)
	lrc := byte(0)
	for i := 0; i < 6; i++ {
		lrc ^= header[i]
	}

	// Include data in LRC if present
	if data != nil && len(data) > 0 {
		for _, b := range data {
			lrc ^= b
		}
	}

	// Build and send packet
	var packet []byte
	packet = append(packet, header...)
	if data != nil && len(data) > 0 {
		packet = append(packet, data...)
	}
	packet = append(packet, lrc)

	written, err := dp.conn.Write(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to write packet: %w", err)
	}
	if written != len(packet) {
		return nil, fmt.Errorf("incomplete write: wrote %d bytes, expected %d", written, len(packet))
	}

	// Read response: wait for sync byte
	syncByte := byte(0)
	for syncByte != ResponseSyncByte {
		buf, err := dp.conn.Read(1)
		if err != nil {
			return nil, fmt.Errorf("failed to read sync byte: %w", err)
		}
		syncByte = buf[0]
	}

	// Read status bytes
	statusBytes, err := dp.conn.Read(2)
	if err != nil {
		return nil, fmt.Errorf("failed to read status bytes: %w", err)
	}
	dp.status0 = statusBytes[0]
	dp.status1 = statusBytes[1]

	// Read data if requested
	var readBytes []byte
	if readLength > 0 {
		readBytes, err = dp.conn.Read(int(readLength))
		if err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}
	}

	// Read LRC byte (but don't verify it for now)
	_, err = dp.conn.Read(1)
	if err != nil {
		return nil, fmt.Errorf("failed to read LRC: %w", err)
	}

	return readBytes, nil
}

// EnterDebug sends the command to make the Foenix enter debug mode
// This stops the CPU and enables debug commands
func (dp *DebugPort) EnterDebug() error {
	_, err := dp.transfer(CMDEnterDebug, 0, nil, 0)
	return err
}

// ExitDebug sends the command to make the Foenix leave debug mode
// This will reset the CPU
func (dp *DebugPort) ExitDebug() error {
	_, err := dp.transfer(CMDExitDebug, 0, nil, 0)
	return err
}

// StopCPU sends the command to stop the CPU from processing instructions (F256 only)
func (dp *DebugPort) StopCPU() error {
	_, err := dp.transfer(CMDStopCPU, 0, nil, 0)
	return err
}

// StartCPU sends the command to restart the CPU after a stop (F256 only)
// This command will not trigger a reset of the CPU
func (dp *DebugPort) StartCPU() error {
	_, err := dp.transfer(CMDStartCPU, 0, nil, 0)
	return err
}

// GetRevision gets the revision code for the debug interface
// RevB2's revision code is 0, RevC4A is 1
func (dp *DebugPort) GetRevision() (byte, error) {
	_, err := dp.transfer(CMDRevision, 0, nil, 0)
	if err != nil {
		return 0, err
	}
	return dp.status1, nil
}

// ReadBlock reads a block of data from the specified address
func (dp *DebugPort) ReadBlock(address uint32, length uint16) ([]byte, error) {
	return dp.transfer(CMDReadMem, address, nil, length)
}

// WriteBlock writes a block of data to the specified address
// For 32-bit 680x0 CPUs (68040/68060), this automatically uses WriteBlock32 for alignment
func (dp *DebugPort) WriteBlock(address uint32, data []byte) error {
	if dp.config.CPUIsM68k32() {
		// For 68040 and 68060, use 32-bit aligned writes
		return dp.WriteBlock32(address, data)
	}

	// For other CPUs, direct write
	_, err := dp.transfer(CMDWriteMem, address, data, 0)
	return err
}

// EraseFlash sends the command to erase the entire flash memory
func (dp *DebugPort) EraseFlash() error {
	_, err := dp.transfer(CMDEraseFlash, 0, nil, 0)
	return err
}

// EraseSector erases an 8KB sector of flash memory
// Note: Sectors are 8KB blocks, but physically erased as two consecutive 4KB blocks
func (dp *DebugPort) EraseSector(sector uint8) error {
	// Erase first 4KB block
	address1 := uint32(sector*2) << 16
	if _, err := dp.transfer(CMDEraseSector, address1, nil, 0); err != nil {
		return fmt.Errorf("failed to erase first 4KB block: %w", err)
	}
	time.Sleep(DelayEraseSector)

	// Erase second 4KB block
	address2 := uint32(sector*2+1) << 16
	if _, err := dp.transfer(CMDEraseSector, address2, nil, 0); err != nil {
		return fmt.Errorf("failed to erase second 4KB block: %w", err)
	}
	time.Sleep(DelayEraseSector)

	return nil
}

// ProgramFlash programs the entire flash memory
// Data should already be loaded in RAM at the specified address
func (dp *DebugPort) ProgramFlash(address uint32) error {
	_, err := dp.transfer(CMDProgramFlash, address, nil, 0)
	return err
}

// ProgramSector programs an 8KB sector of flash memory
// Data should already be loaded into RAM at addresses 0x00000 - 0x02000
func (dp *DebugPort) ProgramSector(sector uint8) error {
	address := uint32(sector*2) << 16
	_, err := dp.transfer(CMDProgramSector, address, nil, 0)
	if err != nil {
		return err
	}
	time.Sleep(DelayProgramSector)
	return nil
}

// SetBootSource sets whether the system should boot from RAM LUTs (0) or Flash LUTs (1)
func (dp *DebugPort) SetBootSource(source byte) error {
	if source == BootSrcRAM {
		_, err := dp.transfer(CMDBootRAM, 0, nil, 0)
		return err
	} else if source == BootSrcFlash {
		_, err := dp.transfer(CMDBootFlash, 0, nil, 0)
		return err
	}
	return fmt.Errorf("invalid boot source: %d (must be 0 for RAM or 1 for Flash)", source)
}
