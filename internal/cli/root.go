// Package cli defines cast's command hierarchy and its command-boundary behavior.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/aryankumar/cast/internal/apperror"
	"github.com/aryankumar/cast/internal/config"
	"github.com/aryankumar/cast/internal/logging"
	"github.com/spf13/cobra"
)

// Options makes the command tree testable and lets release builds supply a version.
type Options struct {
	Version string
	Out     io.Writer
	ErrOut  io.Writer
}

// NewRootCommand builds the cast command tree.
func NewRootCommand(options Options) *cobra.Command {
	if options.Version == "" {
		options.Version = "dev"
	}
	if options.Out == nil {
		options.Out = io.Discard
	}
	if options.ErrOut == nil {
		options.ErrOut = io.Discard
	}

	var debug bool
	cmd := &cobra.Command{
		Use:           "cast",
		Short:         "Share local work and handle small developer tasks",
		Long:          "cast is a local-first command-line utility for sharing, clipboard history, calculations, and unit conversion.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       options.Version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := config.EnsureDir(); err != nil {
				return apperror.Wrap("Could not prepare cast's configuration directory.", "Check that your user configuration directory is writable.", err)
			}
			logger := logging.New(options.ErrOut, debug)
			logger.Debug("starting command", "command", cmd.CommandPath())
			return nil
		},
	}
	cmd.SetOut(options.Out)
	cmd.SetErr(options.ErrOut)
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "show diagnostic logging on stderr")
	cmd.AddCommand(newCalcCommand(), newClipCommand(), newConvCommand(), newReceiveCommand(), newShareCommand(), newVersionCommand(options.Version))
	return cmd
}

func newVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the cast version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}

// PrintError renders normal application failures without leaking raw stack traces.
func PrintError(writer io.Writer, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		fmt.Fprintf(writer, "Error: %s\n", appErr.Message)
		if appErr.Hint != "" {
			fmt.Fprintf(writer, "%s\n", appErr.Hint)
		}
		return
	}
	fmt.Fprintf(writer, "Error: %s\n", err)
}
