package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/atuleu/go-humanize"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/logger"
	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
)

const VERSION string = "0.0.1"

func main() {
	cmd := MakeCliCommand()
	cmd.Run(context.Background(), os.Args)
}

// run takes options and starts the download based on the options given
func run(opts *options.NitroOptions) {
	if strings.HasPrefix(strings.ToLower(opts.Url), "http") {
		metadata, err := metafetcher.FetchMetadataHttp(opts.Url)
		if err != nil {
			panic(err)
		}

		isMetadataProper := metadata.AcceptRanges && metadata.ContentLength > 0

		var pBars *ProgressBars = nil

		if isMetadataProper {
			metadata.LogMetaData()
			pBars, err = MakeProgressBars(opts.Parallel, metadata.ContentLength)
			if err != nil {
				panic(err)
			}
		}

		startTime := time.Now()
		err = downloader.DownloadHttp(
			metadata,
			opts,
			func(partNo uint8, bytesWritten int) {
				if pBars != nil {
					pBars.Bars[partNo].IncrBy(bytesWritten)
				}
			})
		if err != nil {
			panic(err)
		}
		endTime := time.Now()

		if pBars != nil {
			pBars.Container.Wait()
		}

		duration := endTime.Sub(startTime)
		logger.Infof("\nDownloaded in %s\n", humanize.Duration(duration))
	} else if strings.HasPrefix(strings.ToLower(opts.Url), "ftp") {
		metadata, err := metafetcher.FetchMetadataFtp(opts.Url)
		if err != nil {
			panic(err)
		}
		metadata.LogMetaData()

		pBars, err := MakeProgressBars(opts.Parallel, metadata.ContentLength)
		if err != nil {
			panic(err)
		}

		startTime := time.Now()
		err = downloader.DownloadFtp(
			metadata,
			opts,
			func(partNo uint8, bytesWritten int) {
				pBars.Bars[partNo].IncrBy(bytesWritten)
			})
		if err != nil {
			panic(err)
		}
		endTime := time.Now()

		pBars.Container.Wait()

		duration := endTime.Sub(startTime)
		logger.Infof("\nDownloaded in %s\n", humanize.Duration(duration))
	} else {
		panic("Url should start with http or ftp")
	}
}
