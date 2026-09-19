package cli

import (
	"context"
	"fmt"
	"time"

	liveapp "github.com/aryankumar/cast/internal/app/live"
	"github.com/aryankumar/cast/internal/apperror"
	"github.com/spf13/cobra"
)

func newConnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect CODE",
		Short: "Connect to an active live session",
		Long:  "Resolve and connect to an active live project directory or service session on the local network.",
		Args:  cobra.ExactArgs(1),
		RunE:  runConnect,
	}
}

func runConnect(cmd *cobra.Command, args []string) error {
	lookupContext, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	resolved, err := liveapp.Resolve(lookupContext, args[0])
	if err != nil {
		return apperror.Wrap(
			"Could not connect to that live session.",
			"The code may have expired, the live host may be offline, or both devices may not be on the same local network.",
			err,
		)
	}

	writer := cmd.OutOrStdout()
	fmt.Fprintf(writer, "Connected to:\n  %s\n\n", resolved.Device)
	fmt.Fprintf(writer, "Live project:\n  %s\n\n", resolved.Project)
	fmt.Fprintf(writer, "URL:\n  %s\n", resolved.URL)
	return nil
}
