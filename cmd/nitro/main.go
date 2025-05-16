package main

import (
	"fmt"
	"os"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/metafetcher"
)

func main() {
	options, _ := parseCommandLineInputs(os.Args)
	fmt.Println(options)
	metadata := metafetcher.FetchMetadata(options.Url)
	downloader.Download(metadata, options.Parallel)
}
