package cmd

import (
	"fmt"
	"strings"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

// bootCmd represents the boot source command
var bootCmd = &cobra.Command{
	Use:   "boot <source>",
	Short: "Set boot source (ram or flash)",
	Long: `Set the boot source for the Foenix hardware.

For F256jr Rev A and compatible machines, this sets whether the system
should boot from RAM LUTs or Flash LUTs.

Arguments:
  ram   - Boot from RAM lookup tables
  flash - Boot from Flash lookup tables

Example:
  foenixmgr boot ram
  foenixmgr boot flash`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"ram", "flash"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return setBootSource(args[0])
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)
}

// setBootSource sets the boot source to RAM or Flash
func setBootSource(source string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Normalize source
	source = strings.ToLower(source)
	if source != "ram" && source != "flash" {
		return fmt.Errorf("invalid boot source '%s' (must be 'ram' or 'flash')", source)
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

	// Set boot source
	var bootSrc byte
	if source == "ram" {
		printInfo("Setting boot source to RAM...\n")
		bootSrc = protocol.BootSrcRAM
	} else {
		printInfo("Setting boot source to Flash...\n")
		bootSrc = protocol.BootSrcFlash
	}

	if err := dp.SetBootSource(bootSrc); err != nil {
		return fmt.Errorf("failed to set boot source: %w", err)
	}

	printInfo("Boot source set to %s.\n", source)
	return nil
}
