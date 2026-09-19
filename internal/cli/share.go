package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	shareapp "github.com/aryankumar/cast/internal/app/share"
	"github.com/aryankumar/cast/internal/apperror"
	"github.com/aryankumar/cast/internal/qr"
	"github.com/spf13/cobra"
)

const shareLifetime = 10 * time.Minute

func newShareCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "share FILE",
		Short: "Share a file on the local network",
		Long:  "Share a regular file by QR code or short code for ten minutes. The file stays on this device.",
		Args:  cobra.ExactArgs(1),
		RunE:  runShare,
	}
}

func newReceiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "receive CODE",
		Short: "Receive a shared file by code",
		Args:  cobra.ExactArgs(1),
		RunE:  runReceive,
	}
}

func runShare(cmd *cobra.Command, args []string) error {
	server, err := shareapp.StartFile(args[0], shareLifetime)
	if err != nil {
		return apperror.Wrap("Could not start this file share.", "Choose an existing regular file and ensure another cast share is not already running.", err)
	}
	defer server.Stop(context.Background())
	session := server.Session()

	writer := cmd.OutOrStdout()
	fmt.Fprintln(writer, "Sharing:")
	fmt.Fprintf(writer, "  %s\n\n", session.Name)
	fmt.Fprintf(writer, "Size:\n  %s\n\n", formatSize(session.Size))
	fmt.Fprintf(writer, "Code:\n  %s\n\n", session.Code)
	fmt.Fprintln(writer, "Scan the QR code to download:")
	qr.Render(session.URL, writer)
	fmt.Fprintf(writer, "\nOr run:\n  cast receive %s\n\n", session.Code)
	fmt.Fprintln(writer, "Expires in: 10 minutes")
	fmt.Fprintln(writer, "Press Ctrl+C to stop.")

	stopContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-stopContext.Done():
		fmt.Fprintln(writer, "Share stopped.")
	case <-time.After(time.Until(session.ExpiresAt)):
		fmt.Fprintln(writer, "Share expired.")
	}
	return nil
}

func runReceive(cmd *cobra.Command, args []string) error {
	lookupContext, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	resolved, err := shareapp.Resolve(lookupContext, args[0])
	if err != nil {
		return apperror.Wrap("Could not connect to that share.", "The code may have expired, the sharing device may be offline, or both devices may not be on the same local network.", err)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return apperror.Wrap("Could not determine the download location.", "Run cast from a writable directory.", err)
	}
	writer := cmd.OutOrStdout()
	fmt.Fprintf(writer, "Receiving %s...\n", resolved.Name)
	var lastProgress int64 = -1
	path, err := shareapp.Receive(context.Background(), resolved, workingDirectory, func(received, total int64) {
		if total <= 0 {
			return
		}
		percent := received * 100 / total
		if percent != lastProgress {
			lastProgress = percent
			fmt.Fprintf(writer, "\r%d%%  %s / %s", percent, formatSize(received), formatSize(total))
		}
	})
	if err != nil {
		return apperror.Wrap("Could not receive the shared file.", "Check the share code and destination directory; cast will not overwrite an existing file.", err)
	}
	if lastProgress >= 0 {
		fmt.Fprintln(writer)
	}
	fmt.Fprintf(writer, "Received: %s\n", filepath.Base(path))
	return nil
}

func formatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := 0
	for value >= 1000 && unit < len(units)-1 {
		value /= 1000
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}
