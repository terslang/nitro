package main

import (
	"fmt"
	"os"

	"github.com/terslang/nitro/pkg/downloader"
	"github.com/terslang/nitro/pkg/metafetcher"
)

func main() {
	metadata := metafetcher.FetchMetadata(os.Args[1])
	downloader.Download(os.Args[1], metadata)
	fmt.Println(metadata)
}
