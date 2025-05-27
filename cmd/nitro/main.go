package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/helpers"
	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
	"github.com/terslang/nitro/pkg/progressutils"
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

// run takes options and starts the download based on the options given
func run(opts options.NitroOptions) {
	if strings.HasPrefix(strings.ToLower(opts.Url), "http") {
		metadata, err := metafetcher.FetchMetadataHttp(opts.Url)
		if err != nil {
			panic(err)
		}

		segmentTargets := make([]uint64, opts.Parallel)
		partialContentSize, err := helpers.GetPartialContentSize(metadata.ContentLength, opts.Parallel)
		for i := uint8(0); i < opts.Parallel; i++ {
			from, to := helpers.CalculateFromAndToBytes(metadata.ContentLength, partialContentSize, i)
			segmentTargets[i] = to - from + 1
		}

		// Create the segmented progress bar at the top.
		segBar := progressutils.NewSegmentedProgressBar(metadata.ContentLength, segmentTargets, opts.OutputFileName)

		// Define a callback (if needed) to process progress updates.
		progressCallback := func(part int, bytesWritten int) {
			segBar.UpdateSegment(int(part), uint64(bytesWritten))
			fmt.Print("\r", segBar.Render())
		}

		err = downloader.DownloadHttp(metadata, opts, progressCallback)

		if err != nil {
			panic(err)
		}
		segBar.Finish()
		fmt.Println("\nDownload completed successfully!")
	} else if strings.HasPrefix(strings.ToLower(opts.Url), "ftp") {
		metadata, err := metafetcher.FetchMetadataFtp(opts.Url)
		if err != nil {
			panic(err)
		}
		err = downloader.DownloadFtp(metadata, opts)
		if err != nil {
			panic(err)
		}
	} else {
		panic("Url should start with http or ftp")
	}
}
