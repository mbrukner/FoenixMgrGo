// Package cmd implements all CLI commands for FoenixMgr
package cmd

import (
	"fmt"
	"os"

	"github.com/daschewie/foenixmgr/pkg/config"
	"github.com/spf13/cobra"
)

var (
	// Global configuration instance
	cfg *config.Config

	// Global flags
	portFlag   string
	targetFlag string
	quietFlag  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "foenixmgr",
	Short: "FoenixMgr - Manage Foenix retro computers through their debug port",
	Long: `FoenixMgr is a command-line tool for interacting with Foenix retro computers
(C256 Foenix, F256jr, F256k, FNX1591, A2560) through their debug port.

It enables uploading binaries, programming flash memory, reading/writing memory,
and controlling the CPU state over a serial or TCP connection.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Override port from flag if specified
		if portFlag != "" {
			cfg.Port = portFlag
		}

		// Set target machine if specified
		if targetFlag != "" {
			cfg.SetTarget(targetFlag)
		}

		// Quiet mode is handled by printInfo() helper function throughout the codebase
		// (suppresses informational output when quietFlag is true)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&portFlag, "port", "", "Serial port or TCP address (e.g., COM3, /dev/ttyUSB0, 192.168.1.114:2560)")
	rootCmd.PersistentFlags().StringVar(&targetFlag, "target", "", "Target machine (f256jr, f256k, fnx1591, a2560)")
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "Suppress informational output")

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// Helper function to check if connection flags are valid
func validateConnectionFlags() error {
	if cfg.Port == "" && portFlag == "" {
		return fmt.Errorf("no port specified (use --port flag or set in foenixmgr.ini)")
	}
	return nil
}

// Helper function for printing output (respects quiet mode)
func printInfo(format string, args ...interface{}) {
	if !quietFlag {
		fmt.Printf(format, args...)
	}
}

// Helper function for printing errors (always shown)
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
