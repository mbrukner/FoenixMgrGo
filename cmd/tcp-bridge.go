package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/spf13/cobra"
)

// tcpBridgeCmd represents the tcp-bridge command
var tcpBridgeCmd = &cobra.Command{
	Use:   "tcp-bridge <host:port>",
	Short: "Start TCP-to-serial relay server",
	Long: `Start a TCP server that relays debug port protocol messages between
TCP clients and the serial port.

This is useful for:
- Remote development
- macOS systems (driver compatibility)
- Network-based tooling

The TCP server will accept connections on the specified host:port and relay
all debug port protocol messages to the configured serial port.

Example:
  foenixmgr tcp-bridge localhost:2560
  foenixmgr tcp-bridge 0.0.0.0:2560  # Listen on all interfaces`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return startTcpBridge(args[0])
	},
}

func init() {
	rootCmd.AddCommand(tcpBridgeCmd)
}

// startTcpBridge starts the TCP bridge server
func startTcpBridge(hostPort string) error {
	if err := validateConnectionFlags(); err != nil {
		return err
	}

	// Parse host:port
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid host:port format (expected HOST:PORT)")
	}

	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid port number: %w", err)
	}

	printInfo("Starting TCP bridge on %s:%d -> %s\n", host, port, cfg.Port)
	printInfo("Serial settings: %d baud, %d second timeout\n", cfg.DataRate, cfg.Timeout)

	// Create and start bridge
	bridge := connection.NewBridge(host, port, cfg.Port, cfg.DataRate, cfg.Timeout)
	return bridge.Listen()
}
