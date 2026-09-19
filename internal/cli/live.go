package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	liveapp "github.com/aryankumar/cast/internal/app/live"
	"github.com/aryankumar/cast/internal/apperror"
	"github.com/aryankumar/cast/internal/qr"
	"github.com/spf13/cobra"
)

const liveLifetime = 10 * time.Minute

func newLiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "live DIRECTORY|PORT",
		Short: "Make a local project directory accessible on the local network",
		Long: "Serve a project directory as a read-only, expiring local-network session.\n\n" +
			"Examples:\n  cast live .\n  cast live ./project\n  cast live 3000",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if port, err := strconv.Atoi(args[0]); err == nil {
				return runLiveProxy(cmd, port)
			}
			return runLiveDirectory(cmd, args[0])
		},
	}
}

func runLiveProxy(cmd *cobra.Command, port int) error {
	server, err := liveapp.StartProxy(port, liveLifetime)
	if err != nil {
		return apperror.Wrap("Could not start a live service session.", "Ensure a local HTTP service is running on that port, then try again.", err)
	}
	defer server.Stop(context.Background())
	session := server.Session()
	writer := cmd.OutOrStdout()
	fmt.Fprintln(writer, "Casting live:")
	fmt.Fprintf(writer, "\nLocal service:\n  http://localhost:%d\n", port)
	fmt.Fprintf(writer, "\nLocal URL:\n  %s\n", session.LocalURL)
	fmt.Fprintf(writer, "\nRemote URL:\n  %s\n", session.URL)
	fmt.Fprintf(writer, "\nLive code:\n  %s\n\n", session.Code)
	fmt.Fprintln(writer, "Scan the QR code or run:")
	fmt.Fprintf(writer, "  cast connect %s\n\n", session.Code)
	qr.Render(session.URL, writer)
	fmt.Fprintln(writer, "\nExpires in 10 minutes. Press Ctrl+C to stop.")

	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-stopContext.Done():
		fmt.Fprintln(writer, "Live session stopped.")
	case <-time.After(time.Until(session.ExpiresAt)):
		fmt.Fprintln(writer, "Live session expired.")
	}
	return nil
}

func runLiveDirectory(cmd *cobra.Command, root string) error {
	server, err := liveapp.StartDirectory(root, liveLifetime)
	if err != nil {
		return apperror.Wrap("Could not start a live project session.", "Choose an existing directory that cast can read.", err)
	}
	defer server.Stop(context.Background())
	session := server.Session()
	writer := cmd.OutOrStdout()
	fmt.Fprintln(writer, "Casting live:")
	fmt.Fprintf(writer, "\nProject:\n  %s\n", filepath.Clean(root))
	fmt.Fprintf(writer, "\nLocal URL:\n  %s\n", session.LocalURL)
	fmt.Fprintf(writer, "\nRemote URL:\n  %s\n", session.URL)
	fmt.Fprintf(writer, "\nLive code:\n  %s\n\n", session.Code)
	fmt.Fprintln(writer, "Scan the QR code or run:")
	fmt.Fprintf(writer, "  cast connect %s\n\n", session.Code)
	qr.Render(session.URL, writer)
	fmt.Fprintln(writer, "\nExpires in 10 minutes. Press Ctrl+C to stop.")

	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-stopContext.Done():
		fmt.Fprintln(writer, "Live session stopped.")
	case <-time.After(time.Until(session.ExpiresAt)):
		fmt.Fprintln(writer, "Live session expired.")
	}
	return nil
}
