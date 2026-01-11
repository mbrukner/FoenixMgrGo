package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

// copyCmd represents the copy command for F256jr SD card
var copyCmd = &cobra.Command{
	Use:   "copy <file>",
	Short: "Copy file to F256jr SD card",
	Long: `Copy a file to the F256jr SD card.

This command uploads a file to the F256jr's RAM and signals the firmware
to copy it to the SD card. The file is uploaded along with its filename,
size, and CRC32 checksum.

The maximum file size is (7*65536)-(9*1024) bytes (approximately 448 KB).

This command is specific to F256jr hardware.

Example:
  foenixmgr copy program.bin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return copyFile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}

// copyFile copies a file to the F256jr SD card
func copyFile(filename string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Maximum file size: (7*65536)-(9*1024)
	const maxFileSize = (7 * 65536) - (9 * 1024)

	// Read file
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileSize := len(fileData)
	if fileSize >= maxFileSize {
		return fmt.Errorf("file too large (%d bytes, max %d bytes)", fileSize, maxFileSize)
	}

	// Calculate CRC32
	crc32 := util.CalculateCRC32(fileData)

	printInfo("File: %s\n", filename)
	printInfo("Size: %d bytes\n", fileSize)
	printInfo("CRC32: 0x%08X\n", crc32)

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	dp := protocol.NewDebugPort(conn, cfg)

	// Enter debug mode
	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Upload file data to RAM starting at 0x10000
	printInfo("Uploading file data to RAM...\n")

	currentAddr := uint32(0x10000)

	// 1. Write filename (null-terminated)
	basename := filepath.Base(filename)
	filenameBytes := []byte(basename)
	if err := dp.WriteBlock(currentAddr, filenameBytes); err != nil {
		return fmt.Errorf("failed to write filename: %w", err)
	}
	currentAddr += uint32(len(filenameBytes))

	// Null terminator
	if err := dp.WriteBlock(currentAddr, []byte{0x00}); err != nil {
		return fmt.Errorf("failed to write null terminator: %w", err)
	}
	currentAddr++

	// 2. Write CRC32 (4 bytes, little-endian)
	crc32Bytes := []byte{
		byte(crc32 & 0xFF),
		byte((crc32 >> 8) & 0xFF),
		byte((crc32 >> 16) & 0xFF),
		byte((crc32 >> 24) & 0xFF),
	}
	if err := dp.WriteBlock(currentAddr, crc32Bytes); err != nil {
		return fmt.Errorf("failed to write CRC32: %w", err)
	}
	currentAddr += 4

	// 3. Write file size (3 bytes, little-endian)
	sizeBytes := []byte{
		byte(fileSize & 0xFF),
		byte((fileSize >> 8) & 0xFF),
		byte((fileSize >> 16) & 0xFF),
	}
	if err := dp.WriteBlock(currentAddr, sizeBytes); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}
	currentAddr += 3

	// 4. Write file data in chunks
	chunkSize := cfg.ChunkSize
	totalLength := len(fileData)
	chunkOffset := 0

	for totalLength > 0 {
		currentChunkSize := chunkSize
		if totalLength < chunkSize {
			currentChunkSize = totalLength
		}

		chunk := fileData[chunkOffset : chunkOffset+currentChunkSize]
		if err := dp.WriteBlock(currentAddr, chunk); err != nil {
			return fmt.Errorf("failed to write file chunk: %w", err)
		}

		currentAddr += uint32(currentChunkSize)
		totalLength -= currentChunkSize
		chunkOffset += currentChunkSize
	}

	// 5. Trigger firmware copy by writing "COPYFILE" signature to 0x0080
	printInfo("Signaling firmware to copy to SD card...\n")
	copySignature := []byte{0x43, 0x4f, 0x50, 0x59, 0x46, 0x49, 0x4c, 0x45} // "COPYFILE"
	if err := dp.WriteBlock(0x0080, copySignature); err != nil {
		return fmt.Errorf("failed to write copy signature: %w", err)
	}

	printInfo("Copy request sent to firmware.\n")
	return nil
}
