package cmd

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

var labelFile string

// lookupCmd represents the lookup command
var lookupCmd = &cobra.Command{
	Use:   "lookup <label>",
	Short: "Display memory at label's address",
	Long: `Look up a label in the label file and display memory at that address.

The label file is a 64TASS format file with entries like:
  LABEL = $ADDRESS

Example:
  foenixmgr lookup my_variable --label-file program.lbl --count 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return lookupLabel(args[0])
	},
}

// derefCmd represents the dereference command
var derefCmd = &cobra.Command{
	Use:   "deref <label>",
	Short: "Dereference pointer at label and display target memory",
	Long: `Look up a label in the label file, read the 24-bit pointer stored there,
and display memory at the dereferenced address.

This is useful for following pointers in assembly code.

The pointer is read as 3 bytes in little-endian format (6502/65816 style).

Example:
  foenixmgr deref ptr_variable --label-file program.lbl --count 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return derefLabel(args[0])
	},
}

func init() {
	rootCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(derefCmd)

	// Add label-file flag (defaults from config)
	lookupCmd.Flags().StringVar(&labelFile, "label-file", "", "64TASS label file")
	lookupCmd.Flags().StringVar(&dumpCount, "count", "10", "Number of bytes to display (hex)")

	derefCmd.Flags().StringVar(&labelFile, "label-file", "", "64TASS label file")
	derefCmd.Flags().StringVar(&dumpCount, "count", "10", "Number of bytes to display (hex)")
}

// lookupLabel looks up a label and displays memory at that address
func lookupLabel(label string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Determine label file path
	lblFile := labelFile
	if lblFile == "" {
		lblFile = cfg.LabelFile
	}

	// Load label file
	labels := util.NewLabelFile()
	if err := labels.Load(lblFile); err != nil {
		return fmt.Errorf("failed to load label file: %w", err)
	}

	// Look up label
	addressHex, err := labels.Lookup(label)
	if err != nil {
		return err
	}

	// Parse address
	address, err := util.ParseHexAddress(addressHex)
	if err != nil {
		return fmt.Errorf("invalid address for label '%s': %w", label, err)
	}

	// Parse count
	count, err := util.ParseHexSize(dumpCount)
	if err != nil {
		return fmt.Errorf("invalid count: %w", err)
	}

	printInfo("Label '%s' -> Address 0x%X\n", label, address)

	// Create connection and read memory
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	dp := protocol.NewDebugPort(conn, cfg)

	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Read memory
	data, err := dp.ReadBlock(address, count)
	if err != nil {
		return fmt.Errorf("failed to read memory: %w", err)
	}

	// Display hex dump
	util.HexDump(data, address)

	return nil
}

// derefLabel dereferences a pointer at label and displays target memory
func derefLabel(label string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Determine label file path
	lblFile := labelFile
	if lblFile == "" {
		lblFile = cfg.LabelFile
	}

	// Load label file
	labels := util.NewLabelFile()
	if err := labels.Load(lblFile); err != nil {
		return fmt.Errorf("failed to load label file: %w", err)
	}

	// Look up label
	addressHex, err := labels.Lookup(label)
	if err != nil {
		return err
	}

	// Parse address
	address, err := util.ParseHexAddress(addressHex)
	if err != nil {
		return fmt.Errorf("invalid address for label '%s': %w", label, err)
	}

	// Parse count
	count, err := util.ParseHexSize(dumpCount)
	if err != nil {
		return fmt.Errorf("invalid count: %w", err)
	}

	// Create connection
	conn := connection.NewConnection(cfg.Port)
	if err := conn.Open(cfg.Port); err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer conn.Close()

	dp := protocol.NewDebugPort(conn, cfg)

	isStopped := util.IsStopped()
	if !isStopped {
		if err := dp.EnterDebug(); err != nil {
			return fmt.Errorf("failed to enter debug mode: %w", err)
		}
		defer dp.ExitDebug()
	}

	// Read 3 bytes at label's address (24-bit pointer, little-endian)
	printInfo("Label '%s' -> Pointer at 0x%X\n", label, address)

	pointerBytes, err := dp.ReadBlock(address, 3)
	if err != nil {
		return fmt.Errorf("failed to read pointer: %w", err)
	}

	// Decode 24-bit little-endian pointer
	targetAddress := uint32(pointerBytes[0]) |
		uint32(pointerBytes[1])<<8 |
		uint32(pointerBytes[2])<<16

	printInfo("Pointer value: 0x%06X\n", targetAddress)

	// Read memory at target address
	data, err := dp.ReadBlock(targetAddress, count)
	if err != nil {
		return fmt.Errorf("failed to read dereferenced memory: %w", err)
	}

	// Display hex dump
	util.HexDump(data, targetAddress)

	return nil
}
