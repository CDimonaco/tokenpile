package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/pricing"
)

func pricingCommands(loader *pricing.Loader, overridePath string) *cli.Command {
	return &cli.Command{
		Name:  "pricing",
		Usage: "manage model pricing configuration",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "display merged pricing config",
				Action: func(c *cli.Context) error {
					models := loader.All()

					names := make([]string, 0, len(models))
					for name := range models {
						names = append(names, name)
					}

					sort.Strings(names)

					fmt.Fprintf(c.App.Writer, "%-30s %-20s %s\n", "Model", "In ($/M tokens)", "Out ($/M tokens)")
					fmt.Fprintf(
						c.App.Writer,
						"%s\n",
						"------------------------------------------------------------------------",
					)

					for _, name := range names {
						p := models[name]
						fmt.Fprintf(c.App.Writer, "%-30s %-20.4f %.4f\n",
							name, p.InputPerMillion, p.OutputPerMillion)
					}

					return nil
				},
			},
			{
				Name:      "set",
				Usage:     "add or update model pricing",
				ArgsUsage: "<model>",
				Flags: []cli.Flag{
					&cli.Float64Flag{
						Name:     "in",
						Usage:    "input price per million tokens",
						Required: true,
					},
					&cli.Float64Flag{
						Name:     "out",
						Usage:    "output price per million tokens",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					model := c.Args().First()
					if model == "" {
						return errors.New("model name is required")
					}

					inPrice := c.Float64("in")
					outPrice := c.Float64("out")

					if err := loader.SetOverride(overridePath, model, inPrice, outPrice); err != nil {
						return fmt.Errorf("set pricing: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "Updated pricing for %s: in=%.4f out=%.4f (per million tokens)\n",
						model, inPrice, outPrice)

					return nil
				},
			},
		},
	}
}
