package cli

import (
	"fmt"
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
			"Examples:\n  cast calc 15 '*' 32\n  cast calc '(12 + 8) * 3'",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expression := strings.Join(args, " ")
			result, err := calculator.Evaluate(expression)
			if err != nil {
				return apperror.Wrap("Could not evaluate expression.", "Check the syntax around: "+strconv.Quote(expression), err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), strconv.FormatFloat(result, 'f', -1, 64))
			return nil
		},
	}
}
