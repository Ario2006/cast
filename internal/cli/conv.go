package cli

import (
	"fmt"
	"os"

	"github.com/aryankumar/cast/internal/apperror"
	"github.com/aryankumar/cast/internal/ui"
	"github.com/aryankumar/cast/internal/units"
	"github.com/spf13/cobra"
)

func newConvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "conv VALUE UNIT [TARGET_UNIT]",
		Short: "Convert a value to another unit",
		Long: "Convert length, mass, temperature, time, decimal data size, or speed.\n\n" +
			"Without TARGET_UNIT, cast opens an interactive selector. Examples:\n  cast conv 10 km\n  cast conv 10 km mi\n  cast conv 72 mph",
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			quantity, err := units.Parse(args[0], args[1])
			if err != nil {
				return apperror.Wrap("Could not parse that conversion.", "Provide a number followed by a supported unit, such as: cast conv 10 km", err)
			}
			target := ""
			if len(args) == 3 {
				target = args[2]
			} else {
				target, err = selectTarget(cmd, quantity)
				if err != nil {
					return err
				}
				if target == "" {
					return nil
				}
			}
			converted, err := units.Convert(quantity, target)
			if err != nil {
				return apperror.Wrap("Could not convert that value.", "Choose a target unit in the same category.", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s = %s\n", units.Format(quantity), units.Format(converted))
			return nil
		},
	}
}

func selectTarget(cmd *cobra.Command, quantity units.Quantity) (string, error) {
	targets := units.Targets(quantity.Unit)
	choices := make([]string, len(targets))
	for index, target := range targets {
		choices[index] = target.Name + " (" + target.Symbol + ")"
	}
	selection, selected, err := ui.Select(os.Stdin, cmd.OutOrStdout(), units.Format(quantity)+" →", choices)
	if err != nil {
		return "", apperror.Wrap("Could not open the conversion selector.", "Try providing a target unit directly, for example: cast conv 10 km mi", err)
	}
	if !selected {
		return "", nil
	}
	for index, choice := range choices {
		if choice == selection {
			return targets[index].Symbol, nil
		}
	}
	return "", apperror.New("Could not determine the selected unit.", "Try the conversion again.")
}
