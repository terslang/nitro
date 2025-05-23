package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
)

const VERSION string = "0.0.1"

func main() {
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
				Value:       uint8(runtime.NumCPU()),
				DefaultText: fmt.Sprintf("%d - Determined based on hardware capabilities", uint8(runtime.NumCPU())),
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Path to destination file to download content",
				Aliases: []string{"o"},
				Value:   options.DefaultFileName,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var opts options.NitroOptions
			opts.Url = c.StringArg("url")
			opts.Parallel = c.Uint8("parallel")
			opts.OutputFileName = c.String("output")

			if opts.Url == "" {
				fmt.Println("Incorrect Usage: Positional Argument 'URL' is required")
				cli.ShowAppHelp(c)
				return fmt.Errorf("Positional Argument 'URL' is required")
			}

			run(opts)
			return nil
		},
	}

	cmd.Run(context.Background(), os.Args)
}

func run(opts options.NitroOptions) {
	if strings.HasPrefix(strings.ToLower(opts.Url), "http") {
		metadata, err := metafetcher.FetchMetadataHttp(opts.Url)
		if err != nil {
			panic(err)
		}
		err = downloader.DownloadHttp(metadata, opts)
		if err != nil {
			panic(err)
		}
	} else if strings.HasPrefix(strings.ToLower(opts.Url), "ftp") {
		metadata, err := metafetcher.FetchMetadataFtp(opts.Url)
		if err != nil {
			panic(err)
		}
		err = downloader.DownloadTcp(metadata, opts)
		if err != nil {
			panic(err)
		}
	} else {
		panic("Url should start with http or ftp")
	}
}
