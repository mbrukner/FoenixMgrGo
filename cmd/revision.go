package cmd

import (
	"fmt"

	"github.com/daschewie/foenixmgr/pkg/connection"
	"github.com/daschewie/foenixmgr/pkg/protocol"
	"github.com/daschewie/foenixmgr/pkg/util"
	"github.com/spf13/cobra"
)

var revisionCmd = &cobra.Command{
	Use:   "revision",
	Short: "Get debug port revision code",
	Long: `Query the debug port revision code from the Foenix hardware.
RevB2 returns 0, RevC4A returns 1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate connection flags
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

		// Get revision
		rev, err := dp.GetRevision()
		if err != nil {
			return fmt.Errorf("failed to get revision: %w", err)
		}

		// Print revision
		fmt.Printf("%X\n", rev)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(revisionCmd)
}
