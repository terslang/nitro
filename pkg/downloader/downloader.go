package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/terslang/nitro/pkg/metafetcher"
)

func getPartialContentSize(comtentLength uint64, parallel uint8) uint64 {
	if parallel == 0 {
		panic("division by zero")
	}
	return (comtentLength + uint64(parallel) - 1) / uint64(parallel)
}

func Download(metadata metafetcher.MetaData, parallel uint8) {
	outFile, err := os.Create(metadata.FileName)
	if err != nil {
		fmt.Println("failed to create a file")
	}
	defer outFile.Close()

	var wg sync.WaitGroup

	partialContentSize := getPartialContentSize(metadata.ContentLength, parallel)

	for i := uint8(0); i < parallel; i++ {
		wg.Add(1)
		go func(partNo uint8) {
			defer wg.Done()

			fmt.Println("Starting partial download", partNo)

			rangeFromBytes := partialContentSize * uint64(partNo)
			rangeToBytes := rangeFromBytes + partialContentSize - 1

			rangeToBytes = min(rangeToBytes, metadata.ContentLength)

			partialContents, err := downloadPartial(metadata.Url, rangeFromBytes, rangeToBytes)
			if err != nil {
				fmt.Println("Failed to download file")
			}

			fmt.Printf("Part %d: Downloaded %d bytes. Writing to file at offset %d.\n", partNo, len(partialContents), rangeFromBytes)
			bytesWritten, err := outFile.WriteAt(partialContents, int64(rangeFromBytes))
			if err != nil {
				fmt.Printf("Part %d: Failed to write to file at offset %d: %v\n", partNo, rangeFromBytes, err)
				return // Exit this goroutine on write error
			}
			if bytesWritten != len(partialContents) {
				fmt.Printf("Part %d: Warning - partial write. Expected %d bytes, wrote %d bytes.\n", partNo, len(partialContents), bytesWritten)
			}

			fmt.Printf("%d Partial download is done\n", partNo)
		}(i)
	}

	wg.Wait()
}

func downloadPartial(url string, bytesRangeFrom uint64, bytesRangeTo uint64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("failed to create error")
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", bytesRangeFrom, bytesRangeTo))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to download file")
	}

	contents, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return contents, nil
}
