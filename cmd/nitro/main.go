package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
)

const VERSION string = "0.0.1"

func main() {
	opts := options.NitroOptions{}

	cmd := &cli.Command{
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
				DefaultText: fmt.Sprintf("%d - Determined based on hardware capabilities", uint8(runtime.NumCPU())),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			opts.Url = c.StringArg("url")
			opts.Parallel = c.Uint8("parallel")

			if opts.Parallel == 0 {
				opts.Parallel = uint8(runtime.NumCPU())
			}

			run(opts)
			return nil
		},
	}

	cmd.Run(context.Background(), os.Args)
}

func run(opts options.NitroOptions) {
	fmt.Println(opts)
	metadata := metafetcher.FetchMetadata(opts.Url)
	downloader.Download(metadata, opts.Parallel)
}
