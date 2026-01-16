package cmd

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/loader"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

var uploadAddress string

// uploadCmd represents the Intel HEX upload command
var uploadCmd = &cobra.Command{
	Use:   "upload <hexfile>",
	Short: "Upload Intel HEX format file",
	Long: `Upload a program in Intel HEX format to the Foenix hardware.

Example:
  foenixmgr upload program.hex`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFile(args[0], "intelhex")
	},
}

// uploadSrecCmd represents the SREC upload command
var uploadSrecCmd = &cobra.Command{
	Use:   "upload-srec <srecfile>",
	Short: "Upload Motorola SREC format file",
	Long: `Upload a program in Motorola SREC format to the Foenix hardware.

Example:
  foenixmgr upload-srec program.srec`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFile(args[0], "srec")
	},
}

// uploadWdcCmd represents the WDC binary upload command
var uploadWdcCmd = &cobra.Command{
	Use:   "upload-wdc <wdcfile>",
	Short: "Upload WDCTools binary format file",
	Long: `Upload a program in WDCTools binary format to the Foenix hardware.

Example:
  foenixmgr upload-wdc program.bin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFile(args[0], "wdc")
	},
}

// binaryCmd represents the raw binary upload command
var binaryCmd = &cobra.Command{
	Use:   "binary <binfile>",
	Short: "Upload raw binary file to RAM",
	Long: `Upload a raw binary file to the Foenix hardware at the specified address.

Example:
  foenixmgr binary program.bin --address 380000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadBinary(args[0])
	},
}

// runPgxCmd represents the PGX upload and run command
var runPgxCmd = &cobra.Command{
	Use:   "run-pgx <pgxfile>",
	Short: "Upload and run PGX executable",
	Long: `Upload a PGX format executable and configure reset vectors to run on CPU reset.

PGX files include CPU type verification and will fail if the file doesn't match
the configured CPU.

Example:
  foenixmgr run-pgx program.pgx`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFile(args[0], "pgx")
	},
}

// runPgzCmd represents the PGZ upload and run command
var runPgzCmd = &cobra.Command{
	Use:   "run-pgz <pgzfile>",
	Short: "Upload and run PGZ executable",
	Long: `Upload a PGZ format (compressed) executable and configure reset vectors.

PGZ files can contain multiple data blocks and start address information.

Example:
  foenixmgr run-pgz program.pgz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadFile(args[0], "pgz")
	},
}

// runM68kBinCmd represents the 68k binary upload command
var runM68kBinCmd = &cobra.Command{
	Use:   "run-m68k-bin <binfile>",
	Short: "Upload Motorola 68k binary and set reset vector",
	Long: `Upload a Motorola 68k binary file to RAM and configure the reset vector.

The binary is uploaded to the specified address, and the first 8 bytes
(initial stack pointer and reset vector) are copied to address 0, allowing
the program to start when the CPU exits debug mode.

Example:
  foenixmgr run-m68k-bin program.bin --address 380000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return uploadM68kBinary(args[0])
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(uploadSrecCmd)
	rootCmd.AddCommand(uploadWdcCmd)
	rootCmd.AddCommand(binaryCmd)
	rootCmd.AddCommand(runPgxCmd)
	rootCmd.AddCommand(runPgzCmd)
	rootCmd.AddCommand(runM68kBinCmd)

	// Add --address flag to commands that need it
	binaryCmd.Flags().StringVar(&uploadAddress, "address", "", "Target address (hex, e.g., 380000)")
	binaryCmd.MarkFlagRequired("address")

	runM68kBinCmd.Flags().StringVar(&uploadAddress, "address", "", "Target address (hex, e.g., 380000)")
	runM68kBinCmd.MarkFlagRequired("address")
}

// uploadFile is the common upload handler for all file formats
func uploadFile(filename string, format string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
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

	// Create appropriate loader
	var ldr loader.Loader
	switch format {
	case "intelhex":
		ldr = loader.NewIntelHexLoader()
	case "srec":
		ldr = loader.NewSRecLoader()
	case "wdc":
		ldr = loader.NewWDCLoader()
	case "pgx":
		ldr = loader.NewPGXLoader(cfg)
	case "pgz":
		ldr = loader.NewPGZLoader(cfg)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Open file
	if err := ldr.Open(filename); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer ldr.Close()

	// Set handler to write to debug port
	ldr.SetHandler(func(address uint32, data []byte) error {
		return dp.WriteBlock(address, data)
	})

	// Process file
	printInfo("Uploading %s...\n", filename)
	if err := ldr.Process(); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	printInfo("Upload complete.\n")
	return nil
}

// uploadBinary uploads a raw binary file to the specified address
func uploadBinary(filename string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Parse address
	addr, err := util.ParseHexAddress(uploadAddress)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Read binary file
	data, err := util.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
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

	// Upload binary in chunks (matching Python behavior)
	printInfo("Uploading %d bytes to 0x%X...\n", len(data), addr)
	chunkSize := cfg.ChunkSize
	for offset := 0; offset < len(data); offset += chunkSize {
		end := offset + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]
		if err := dp.WriteBlock(addr+uint32(offset), chunk); err != nil {
			return fmt.Errorf("upload failed at offset 0x%X: %w", offset, err)
		}
	}

	printInfo("Upload complete.\n")
	return nil
}

// uploadM68kBinary uploads a 68k binary and sets up reset vectors
func uploadM68kBinary(filename string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Parse address
	addr, err := util.ParseHexAddress(uploadAddress)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Read binary file
	data, err := util.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Verify file has at least 8 bytes (for stack pointer + reset vector)
	if len(data) < 8 {
		return fmt.Errorf("binary file too small (need at least 8 bytes for vectors)")
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

	// Upload binary to target address in chunks
	printInfo("Uploading %d bytes to 0x%X...\n", len(data), addr)
	chunkSize := cfg.ChunkSize
	for offset := 0; offset < len(data); offset += chunkSize {
		end := offset + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]
		if err := dp.WriteBlock(addr+uint32(offset), chunk); err != nil {
			return fmt.Errorf("upload failed at offset 0x%X: %w", offset, err)
		}
	}

	// Copy first 8 bytes (initial SP and reset vector) to address 0
	printInfo("Setting up reset vectors at address 0...\n")
	if err := dp.WriteBlock(0, data[0:8]); err != nil {
		return fmt.Errorf("failed to set reset vectors: %w", err)
	}

	printInfo("Upload complete. Binary will start at 0x%X on CPU reset.\n", addr)
	return nil
}
