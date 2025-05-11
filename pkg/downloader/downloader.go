package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/terslang/nitro/pkg/metafetcher"
)

func Download(url string, metadata metafetcher.MetaData) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to download file")
	}

	outFile, err := os.Create(metadata.FileName)
	if err != nil {
		fmt.Println("failed to create a file")
	}
	defer outFile.Close()

	bytesWritten, err := io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Println("failed to copy contents")
	}

	fmt.Println("Downloaded file. No of bytes written ", bytesWritten)

}
