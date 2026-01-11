package cmd

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

var (
	dumpAddress string
	dumpCount   string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Read and display memory from specified address",
	Long: `Read a block of memory from the Foenix hardware and display it in hex dump format.

Example:
  foenixmgr dump --address 380000 --count 100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags
		if err := validateConnectionFlags(); err != nil {
			return err
		}

		if dumpAddress == "" {
			// Use default address from config
			dumpAddress = cfg.Address
		}

		if dumpCount == "" {
			dumpCount = "10" // Default to 16 bytes (0x10)
		}

		// Parse address and count
		addr, err := util.ParseHexAddress(dumpAddress)
		if err != nil {
			return fmt.Errorf("invalid address: %w", err)
		}

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

		// Read memory
		data, err := dp.ReadBlock(addr, count)
		if err != nil {
			return fmt.Errorf("failed to read memory: %w", err)
		}

		// Display hex dump
		util.HexDump(data, addr)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	dumpCmd.Flags().StringVar(&dumpAddress, "address", "", "Starting address (hex, e.g., 380000)")
	dumpCmd.Flags().StringVar(&dumpCount, "count", "10", "Number of bytes to read (hex, e.g., 100)")
}
