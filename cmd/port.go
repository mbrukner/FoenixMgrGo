package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

// listPortsCmd represents the list-ports command
var listPortsCmd = &cobra.Command{
	Use:   "list-ports",
	Short: "List available serial ports",
	Long: `List all available serial ports on the system.

This helps identify which port to use for connecting to your Foenix hardware.

Example:
  foenixmgr list-ports`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPorts()
	},
}

func init() {
	rootCmd.AddCommand(listPortsCmd)
}

// listPorts lists all available serial ports
func listPorts() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return fmt.Errorf("failed to get port list: %w", err)
	}

	if len(ports) == 0 {
		fmt.Println("No serial ports found")
		return nil
	}

	fmt.Println("Available serial ports:")
	for _, port := range ports {
		fmt.Printf("  %s\n", port)
	}

	return nil
}
