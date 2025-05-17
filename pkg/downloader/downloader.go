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

			partialContentsReader, err := downloadPartial(metadata.Url, rangeFromBytes, rangeToBytes)
			if err != nil {
				fmt.Println("Failed to download file")
			}

			downloadAndWriteToFile(partialContentsReader, outFile, int64(rangeFromBytes), partNo)

			fmt.Printf("%d Partial download is done\n", partNo)
		}(i)
	}

	wg.Wait()
}

func downloadPartial(url string, bytesRangeFrom uint64, bytesRangeTo uint64) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("failed to create error")
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", bytesRangeFrom, bytesRangeTo))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to download file")
	}

	return resp.Body, nil
}

func downloadAndWriteToFile(reader io.ReadCloser, file *os.File, startOffset int64, partNo uint8) error {
	defer reader.Close()

	buffer := make([]byte, 32*1024) // 32kb buffer
	currentWriteOffset := startOffset

	for {
		bytesRead, readErr := reader.Read(buffer)

		if bytesRead > 0 {
			bytesWritten, writeErr := file.WriteAt(buffer[:bytesRead], currentWriteOffset)
			if writeErr != nil {
				return fmt.Errorf("part %d: failed to write to file at offset %d: %w", partNo, currentWriteOffset, writeErr)
			}

			fmt.Printf("Part %d: Downloaded to file at offset %d\n", partNo, currentWriteOffset)

			currentWriteOffset += int64(bytesWritten)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			return fmt.Errorf("part %d: failed to read response stream: %w", partNo, readErr)
		}
	}

	return nil
}
