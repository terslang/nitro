package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/helpers"
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
			helpers.Verbose = c.Bool("verbose")

			if opts.Url == "" {
				helpers.Infoln("Incorrect Usage: Positional Argument 'URL' is required")
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
		metadata.LogMetaData()

		var noOfPBars uint8
		if !metadata.AcceptRanges || metadata.ContentLength == 0 {
			noOfPBars = 1
		} else {
			noOfPBars = opts.Parallel
		}

		pBars, err := makeProgressBars(noOfPBars, metadata.ContentLength)
		if err != nil {
			panic(err)
		}

		err = downloader.DownloadHttp(
			metadata,
			opts,
			func(partNo uint8, bytesWritten int) {
				pBars.Bars[partNo].IncrBy(bytesWritten)
			})
		if err != nil {
			panic(err)
		}

		pBars.Container.Wait()
	} else if strings.HasPrefix(strings.ToLower(opts.Url), "ftp") {
		metadata, err := metafetcher.FetchMetadataFtp(opts.Url)
		if err != nil {
			panic(err)
		}
		metadata.LogMetaData()

		pBars, err := makeProgressBars(opts.Parallel, metadata.ContentLength)
		if err != nil {
			panic(err)
		}

		err = downloader.DownloadFtp(
			metadata,
			opts,
			func(partNo uint8, bytesWritten int) {
				pBars.Bars[partNo].IncrBy(bytesWritten)
			})
		if err != nil {
			panic(err)
		}

		pBars.Container.Wait()
	} else {
		panic("Url should start with http or ftp")
	}
}

type progressBars struct {
	Container *mpb.Progress
	Bars      []*mpb.Bar
}

func makeProgressBars(noOfPBars uint8, contentLength uint64) (*progressBars, error) {
	progress := mpb.New(mpb.WithAutoRefresh())

	partialContentSize, err := helpers.GetPartialContentSize(contentLength, noOfPBars)
	if err != nil {
		return nil, fmt.Errorf("error while making progress bars: %w", err)
	}

	bars := make([]*mpb.Bar, noOfPBars)

	for i := uint8(0); i < noOfPBars; i++ {
		rangeFromBytes, rangeToBytes := helpers.CalculateFromAndToBytes(contentLength, partialContentSize, i)
		barSize := int64(rangeToBytes - rangeFromBytes)
		bar := progress.New(
			barSize,
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("Part %d/%d", i+1, noOfPBars), decor.WC{W: 12, C: decor.DSyncWidth}),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 5}),
				decor.CountersKibiByte("%.2f/%.2f", decor.WC{W: 20}),
			),
		)

		bars[i] = bar
	}

	return &progressBars{
		Container: progress,
		Bars:      bars,
	}, nil
}
