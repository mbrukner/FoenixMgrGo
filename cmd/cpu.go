package cmd

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

// stopCmd represents the CPU stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop CPU execution (F256 only)",
	Long: `Stop the CPU from processing instructions.

This command is specific to F256 machines and puts the CPU into a stopped state.
The CPU will remain stopped until a 'start' command is issued.

This creates a persistent stopped state tracked by the f256.stp file, allowing
multiple debug operations without CPU reset between commands.

Example:
  foenixmgr stop`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopCPU()
	},
}

// startCmd represents the CPU start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start CPU execution (F256 only)",
	Long: `Restart the CPU after a stop command.

This command resumes CPU execution without triggering a reset. The CPU will
continue from where it was stopped.

This clears the persistent stopped state (f256.stp file).

Example:
  foenixmgr start`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startCPU()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(startCmd)
}

// stopCPU stops the CPU and sets the stop indicator
func stopCPU() error {
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

	// Enter debug mode first
	if err := dp.EnterDebug(); err != nil {
		return fmt.Errorf("failed to enter debug mode: %w", err)
	}

	// Stop the CPU
	printInfo("Stopping CPU...\n")
	if err := dp.StopCPU(); err != nil {
		return fmt.Errorf("failed to stop CPU: %w", err)
	}

	// Set the stop indicator file
	if err := util.SetStopIndicator(); err != nil {
		return fmt.Errorf("failed to set stop indicator: %w", err)
	}

	printInfo("CPU stopped. Use 'start' command to resume.\n")
	return nil
}

// startCPU starts the CPU and clears the stop indicator
func startCPU() error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Check if CPU is actually stopped
	if !util.IsStopped() {
		printInfo("CPU is not in stopped state.\n")
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

	// Start the CPU (no need to enter debug mode, we're already in it)
	printInfo("Starting CPU...\n")
	if err := dp.StartCPU(); err != nil {
		return fmt.Errorf("failed to start CPU: %w", err)
	}

	// Exit debug mode
	if err := dp.ExitDebug(); err != nil {
		return fmt.Errorf("failed to exit debug mode: %w", err)
	}

	// Clear the stop indicator file
	if err := util.ClearStopIndicator(); err != nil {
		return fmt.Errorf("failed to clear stop indicator: %w", err)
	}

	printInfo("CPU started and running.\n")
	return nil
}
