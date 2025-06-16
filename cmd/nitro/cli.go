package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/terslang/nitro/pkg/logger"
	"github.com/terslang/nitro/pkg/options"
)

func MakeCliCommand() *cli.Command {
	return &cli.Command{
		Name:    "nitro",
		Version: VERSION,
		Usage:   "Download Accelerator",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "url",
				UsageText: "Download url",
			},
		},
		Flags: []cli.Flag{
			&cli.Uint8Flag{
				Name:        "parallel",
				Usage:       "Number of conccurent connections to use for parallel downloads",
				Aliases:     []string{"p"},
				Value:       uint8(runtime.NumCPU()),
				DefaultText: fmt.Sprintf("%d - Determined based on hardware capabilities", uint8(runtime.NumCPU())),
				Validator: func(parallel uint8) error {
					if parallel == 0 {
						return fmt.Errorf("Parallel cannot be '0'")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Path to destination file to download content",
				Aliases: []string{"o"},
				Value:   options.DefaultFileName,
				Validator: func(output string) error {
					if strings.TrimSpace(output) == "" {
						return fmt.Errorf("Output filename cannot be empty")
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Prints debug logs if set to true",
				Value: false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opts options.NitroOptions
			opts.Url = c.StringArg("url")
			opts.Parallel = c.Uint8("parallel")
			opts.OutputFileName = c.String("output")
			logger.Verbose = c.Bool("verbose")

			if strings.TrimSpace(opts.Url) == "" {
				logger.Infoln("Incorrect Usage: Positional Argument 'URL' is required")
				cli.ShowAppHelp(c)
				return fmt.Errorf("Positional Argument 'URL' is required")
			}

			run(&opts)
			return nil
		},
	}
}
