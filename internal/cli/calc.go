package cli

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aryankumar/cast/internal/app/calculator"
	"github.com/aryankumar/cast/internal/apperror"
	"github.com/spf13/cobra"
)

func newCalcCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "calc EXPRESSION",
		Short: "Evaluate an arithmetic expression",
		Long: "Evaluate decimal arithmetic with +, -, *, /, %, ^, parentheses, and unary minus.\n\n" +
			"Examples:\n  cast calc '15 + 4 * 3'\n  cast calc '(12 + 8) * 3'",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			normalizedArgs := normalizeCalcArgs(args)
			expression := strings.Join(normalizedArgs, " ")
			result, err := calculator.Evaluate(expression)
			if err != nil {
				return apperror.Wrap("Could not evaluate expression.", "Check the syntax around: "+strconv.Quote(expression), err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), strconv.FormatFloat(result, 'f', -1, 64))
			return nil
		},
	}
}

// normalizeCalcArgs detects and recovers from shell globbing where an unquoted '*'
// operator gets expanded into the list of non-hidden files in the current directory.
func normalizeCalcArgs(args []string) []string {
	allMatches, err := filepath.Glob("*")
	if err != nil || len(allMatches) == 0 {
		return args
	}
	var matches []string
	for _, m := range allMatches {
		if !strings.HasPrefix(filepath.Base(m), ".") {
			matches = append(matches, m)
		}
	}
	if len(matches) == 0 {
		return args
	}
	if len(matches) == 1 && isNumeric(matches[0]) {
		return args
	}
	matchLen := len(matches)
	var result []string
	for i := 0; i < len(args); {
		if i+matchLen <= len(args) && sliceEqual(args[i:i+matchLen], matches) {
			result = append(result, "*")
			i += matchLen
		} else {
			result = append(result, args[i])
			i++
		}
	}
	return result
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
