package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	history "github.com/aryankumar/cast/internal/app/clipboard"
	"github.com/aryankumar/cast/internal/apperror"
	"github.com/aryankumar/cast/internal/config"
	platform "github.com/aryankumar/cast/internal/platform/clipboard"
	"github.com/aryankumar/cast/internal/storage"
	"github.com/aryankumar/cast/internal/ui"
	"github.com/spf13/cobra"
)

const clipboardHistoryLimit = 200

func newClipCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "clip",
		Short: "Browse and restore local clipboard history",
		Long: "Browse local plain-text clipboard history and restore an item with the keyboard.\n\n" +
			"History stays on this device and is never synced or logged.",
		Args: cobra.NoArgs,
		RunE: runClip,
	}
	command.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear local clipboard history",
		Args:  cobra.NoArgs,
		RunE:  runClipClear,
	})
	return command
}

func runClip(cmd *cobra.Command, _ []string) error {
	service, closeStore, err := openClipboardService()
	if err != nil {
		return err
	}
	defer closeStore()

	_, skippedSensitive, err := service.Capture()
	if err != nil {
		return apperror.Wrap("Could not read the system clipboard.", "Grant clipboard access if prompted, then try again.", err)
	}
	if skippedSensitive {
		fmt.Fprintln(cmd.OutOrStdout(), "Current clipboard content was not stored because it appears sensitive.")
	}

	watchContext, stopWatching := context.WithCancel(context.Background())
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		_ = service.Watch(watchContext, time.Second)
	}()
	defer func() {
		stopWatching()
		<-watchDone
	}()

	entries, err := service.List(clipboardHistoryLimit)
	if err != nil {
		return apperror.Wrap("Could not load clipboard history.", "Try running: cast clip clear", err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No clipboard history yet. Copy text, then run cast clip again.")
		return nil
	}

	labels := make([]string, len(entries))
	for index, entry := range entries {
		labels[index] = clipboardLabel(entry.Content)
	}
	selected, confirmed, err := ui.SelectIndex(os.Stdin, cmd.OutOrStdout(), "Clipboard History", labels)
	if err != nil {
		return apperror.Wrap("Could not open the clipboard picker.", "Run cast clip from an interactive terminal.", err)
	}
	if !confirmed {
		return nil
	}
	if err := service.Restore(entries[selected]); err != nil {
		return apperror.Wrap("Could not restore the selected clipboard entry.", "Grant clipboard access if prompted, then try again.", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Copied selected clipboard entry.")
	return nil
}

func runClipClear(cmd *cobra.Command, _ []string) error {
	service, closeStore, err := openClipboardService()
	if err != nil {
		return err
	}
	defer closeStore()
	if err := service.Clear(); err != nil {
		return apperror.Wrap("Could not clear clipboard history.", "Check that cast's configuration directory is writable.", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Clipboard history cleared.")
	return nil
}

func openClipboardService() (*history.Service, func(), error) {
	path, err := config.DataPath("history.db")
	if err != nil {
		return nil, nil, apperror.Wrap("Could not locate clipboard history.", "Check that your user configuration directory is available.", err)
	}
	store, err := storage.OpenClipboardStore(path, clipboardHistoryLimit)
	if err != nil {
		return nil, nil, apperror.Wrap("Could not open clipboard history.", "Check that cast's configuration directory is writable.", err)
	}
	return history.NewService(store, platform.System{}), func() { _ = store.Close() }, nil
}

func clipboardLabel(content string) string {
	label := strings.Join(strings.Fields(content), " ")
	const maximumRunes = 76
	if utf8.RuneCountInString(label) <= maximumRunes {
		return label
	}
	runes := []rune(label)
	return string(runes[:maximumRunes-1]) + "…"
}
