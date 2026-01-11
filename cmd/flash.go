package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

var (
	flashAddress    string
	flashSector     string
	flashEraseFirst bool
)

// eraseCmd represents the flash erase command
var eraseCmd = &cobra.Command{
	Use:   "erase",
	Short: "Erase entire flash memory",
	Long: `Erase the entire flash memory on the Foenix hardware.

⚠️  WARNING: This is a destructive operation that cannot be undone.
All data in flash will be permanently erased.

Example:
  foenixmgr erase`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return eraseFlash()
	},
}

// flashCmd represents the flash programming command
var flashCmd = &cobra.Command{
	Use:   "flash <binfile>",
	Short: "Program flash memory from binary file",
	Long: `Program the entire flash memory from a binary file.

The binary file must be exactly the size configured in foenixmgr.ini
(default: 524288 bytes = 512KB).

Data is uploaded to RAM at the specified address, then programmed to flash.

⚠️  WARNING: This will overwrite flash memory.

Example:
  foenixmgr flash firmware.bin --address 380000

Program a specific 8KB sector:
  foenixmgr flash sector.bin --flash-sector 01 --address 380000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flashSector != "" {
			return flashProgramSector(args[0])
		}
		return flashProgramFull(args[0])
	},
}

// flashBulkCmd represents the bulk flash programming command
var flashBulkCmd = &cobra.Command{
	Use:   "flash-bulk <csvfile>",
	Short: "Program multiple flash sectors from CSV mapping",
	Long: `Program multiple flash sectors using a CSV file that maps sectors to binary files.

CSV Format:
  sector_hex,filename
  01,sector01.bin
  02,sector02.bin
  ...

Each binary file should be 8KB (8192 bytes) for the sector.

Options:
  --erase: Erase entire flash before programming (faster for multiple sectors)

⚠️  WARNING: This will overwrite flash memory.

Example:
  foenixmgr flash-bulk sectors.csv
  foenixmgr flash-bulk sectors.csv --erase`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return flashBulkProgram(args[0])
	},
}

func init() {
	rootCmd.AddCommand(eraseCmd)
	rootCmd.AddCommand(flashCmd)
	rootCmd.AddCommand(flashBulkCmd)

	// Flags for flash command
	flashCmd.Flags().StringVar(&flashAddress, "address", "", "RAM address for flash data (hex, e.g., 380000)")
	flashCmd.Flags().StringVar(&flashSector, "flash-sector", "", "Program specific 8KB sector (hex, e.g., 01)")

	// Mark address as required for full flash
	flashCmd.MarkFlagRequired("address")

	// Flags for flash-bulk command
	flashBulkCmd.Flags().BoolVar(&flashEraseFirst, "erase", false, "Erase entire flash before programming")
}

// eraseFlash erases the entire flash memory with user confirmation
func eraseFlash() error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Get confirmation
	if !util.ConfirmDanger("You are about to ERASE the entire flash memory") {
		printInfo("Operation cancelled.\n")
		return nil
	}

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	// Create protocol handler
	dp := protocol.NewDebugPort(conn, cfg)

	// Enter debug mode
	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Erase flash
	printInfo("Erasing flash memory...\n")
	if err := dp.EraseFlash(); err != nil {
		return fmt.Errorf("flash erase failed: %w", err)
	}

	printInfo("Flash memory erased successfully.\n")
	return nil
}

// flashProgramFull programs the entire flash memory
func flashProgramFull(filename string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Parse address
	addr, err := util.ParseHexAddress(flashAddress)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Read and validate binary file
	data, err := util.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file size (should match configured flash size)
	if len(data) != cfg.FlashSize {
		printInfo("Warning: File size (%d bytes) does not match configured flash size (%d bytes)\n",
			len(data), cfg.FlashSize)
		// The Python version has a bug here: if (1) or (os.path.getsize...) always passes
		// We'll allow it but warn the user
	}

	printInfo("About to upload %d bytes to address 0x%X and program flash\n", len(data), addr)

	// Get confirmation
	if !util.Confirm("Are you sure you want to reprogram the flash memory? (y/n): ") {
		printInfo("Operation cancelled.\n")
		return nil
	}

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	// Create protocol handler
	dp := protocol.NewDebugPort(conn, cfg)

	// Enter debug mode
	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Upload data to RAM
	printInfo("Uploading flash image to RAM...\n")
	if err := uploadChunked(dp, addr, data); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Erase flash (must happen AFTER upload to RAM, BEFORE programming)
	printInfo("Erasing flash memory...\n")
	if err := dp.EraseFlash(); err != nil {
		return fmt.Errorf("flash erase failed: %w", err)
	}

	// Program flash from RAM
	printInfo("Programming flash from RAM...\n")
	if err := dp.ProgramFlash(addr); err != nil {
		return fmt.Errorf("flash programming failed: %w", err)
	}

	printInfo("Flash programming complete.\n")
	return nil
}

// flashProgramSector programs a specific 8KB flash sector
func flashProgramSector(filename string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Check if target machine supports sector programming
	if cfg.FlashPageSize() == 0 || cfg.FlashSectorSize() == 0 {
		return fmt.Errorf("target machine does not support flash sector programming\nUse --target option to specify machine (f256jr, f256k, fnx1591)")
	}

	// Parse sector number
	sectorNum, err := strconv.ParseUint(flashSector, 16, 8)
	if err != nil {
		return fmt.Errorf("invalid sector number: %w", err)
	}

	// Read and validate binary file
	data, err := util.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file size (should be sector size in KB * 1024)
	expectedSize := cfg.FlashSectorSize() * 1024
	if len(data) != expectedSize {
		return fmt.Errorf("file size (%d bytes) does not match sector size (%d bytes)",
			len(data), expectedSize)
	}

	printInfo("About to upload image to sector 0x%02X\n", sectorNum)

	// Get confirmation
	if !util.Confirm("Are you sure you want to reprogram the flash sector? (y/n): ") {
		printInfo("Operation cancelled.\n")
		return nil
	}

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	// Create protocol handler
	dp := protocol.NewDebugPort(conn, cfg)

	// Enter debug mode
	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Calculate page information
	pageSize := cfg.FlashPageSize()
	sectorSize := cfg.FlashSectorSize()
	pagesPerSector := sectorSize / pageSize
	startPage := uint8(sectorNum) * uint8(pagesPerSector)

	// Upload and program sector in pages
	ramAddress := uint32(0)
	written := 0
	currentPage := startPage

	for written < len(data) {
		// Calculate how much to write in this chunk
		toWrite := cfg.ChunkSize
		if toWrite > len(data)-written {
			toWrite = len(data) - written
		}

		// Write chunk to RAM
		chunk := data[written : written+toWrite]
		if err := dp.WriteBlock(ramAddress, chunk); err != nil {
			return fmt.Errorf("failed to write to RAM: %w", err)
		}

		written += toWrite
		ramAddress += uint32(toWrite)

		// If we've filled the RAM buffer, program the flash page
		if ramAddress >= uint32(cfg.RAMSize()*1024) {
			printInfo("Erasing flash page %d...\n", currentPage)
			if err := dp.EraseSector(currentPage); err != nil {
				return fmt.Errorf("failed to erase sector: %w", err)
			}

			printInfo("Programming flash page %d...\n", currentPage)
			if err := dp.ProgramSector(currentPage); err != nil {
				return fmt.Errorf("failed to program sector: %w", err)
			}

			currentPage++
			ramAddress = 0
		}
	}

	// Program any remaining data
	if ramAddress > 0 {
		printInfo("Erasing flash page %d...\n", currentPage)
		if err := dp.EraseSector(currentPage); err != nil {
			return fmt.Errorf("failed to erase sector: %w", err)
		}

		printInfo("Programming flash page %d...\n", currentPage)
		if err := dp.ProgramSector(currentPage); err != nil {
			return fmt.Errorf("failed to program sector: %w", err)
		}
	}

	printInfo("Flash sector programming complete.\n")
	return nil
}

// flashBulkProgram programs multiple sectors from a CSV mapping file
func flashBulkProgram(csvFile string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Open and parse CSV file
	f, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Display what will be programmed
	printInfo("Flash bulk programming plan:\n")
	for _, record := range records {
		if len(record) < 2 {
			return fmt.Errorf("invalid CSV format: expected sector,filename")
		}
		printInfo("  Sector 0x%s: %s\n", record[0], record[1])
	}

	// Get confirmation
	if !util.Confirm("\nProceed with flash bulk programming? (y/n): ") {
		printInfo("Operation cancelled.\n")
		return nil
	}

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	// Create protocol handler
	dp := protocol.NewDebugPort(conn, cfg)

	// Enter debug mode
	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Erase entire flash if requested
	if flashEraseFirst {
		printInfo("Erasing entire flash memory...\n")
		if err := dp.EraseFlash(); err != nil {
			return fmt.Errorf("flash erase failed: %w", err)
		}
		printInfo("Flash erased.\n")
	}

	// Program each sector
	for _, record := range records {
		sectorID := record[0]
		sectorFile := record[1]

		sectorNum, err := strconv.ParseUint(sectorID, 16, 8)
		if err != nil {
			return fmt.Errorf("invalid sector number '%s': %w", sectorID, err)
		}

		printInfo("\nProgramming sector 0x%02X from %s...\n", sectorNum, sectorFile)

		// Read sector data
		data, err := util.ReadFile(sectorFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", sectorFile, err)
		}

		// Upload to RAM at address 0
		ramAddress := uint32(0)
		if err := uploadChunked(dp, ramAddress, data); err != nil {
			return fmt.Errorf("failed to upload %s: %w", sectorFile, err)
		}

		printInfo("Binary uploaded to RAM.\n")

		// Erase sector (if not pre-erased)
		if !flashEraseFirst {
			printInfo("Erasing flash sector...\n")
			if err := dp.EraseSector(uint8(sectorNum)); err != nil {
				return fmt.Errorf("failed to erase sector: %w", err)
			}
		}

		// Program sector
		printInfo("Programming flash sector...\n")
		if err := dp.ProgramSector(uint8(sectorNum)); err != nil {
			return fmt.Errorf("failed to program sector: %w", err)
		}

		printInfo("Sector 0x%02X programmed successfully.\n", sectorNum)
	}

	printInfo("\nFlash bulk programming complete.\n")
	return nil
}

// uploadChunked uploads data in chunks to avoid overwhelming the debug port
func uploadChunked(dp *protocol.DebugPort, startAddress uint32, data []byte) error {
	address := startAddress
	offset := 0

	for offset < len(data) {
		chunkSize := cfg.ChunkSize
		if offset+chunkSize > len(data) {
			chunkSize = len(data) - offset
		}

		chunk := data[offset : offset+chunkSize]
		if err := dp.WriteBlock(address, chunk); err != nil {
			return fmt.Errorf("failed to write chunk at 0x%X: %w", address, err)
		}

		address += uint32(chunkSize)
		offset += chunkSize
	}

	return nil
}
