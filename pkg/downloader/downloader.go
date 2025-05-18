package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
)

func getPartialContentSize(comtentLength uint64, parallel uint8) uint64 {
	if parallel == 0 {
		panic("division by zero")
	}
	return (comtentLength + uint64(parallel) - 1) / uint64(parallel)
}

func Download(metadata metafetcher.MetaData, opts options.NitroOptions) error {
	var fileName string
	var parallel uint8

	if opts.OutputFileName == options.DefaultFileName {
		fileName = metadata.FileName
	} else {
		fileName = opts.OutputFileName
	}

	if !metadata.AcceptRanges || metadata.ContentLength == 0 {
		fmt.Println("Download doesn't support partial downloads. Downloading with 1 connection")
		parallel = 1
	} else {
		parallel = opts.Parallel
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Failed to create file. %w", err)
	}
	defer outFile.Close()

	var wg sync.WaitGroup
	errorChannel := make(chan error, parallel)

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
				err = fmt.Errorf("Part %d (range %d-%d): failed to initiate download: %w", partNo, rangeFromBytes, rangeToBytes, err)
				errorChannel <- err
				return
			}

			err = downloadAndWriteToFile(partialContentsReader, outFile, int64(rangeFromBytes), partNo)
			if err != nil {
				err = fmt.Errorf("Part %d (range %d-%d): failed to download: %w", partNo, rangeFromBytes, rangeToBytes, err)
				errorChannel <- err
				return
			}

			fmt.Printf("%d Partial download is done\n", partNo)
		}(i)
	}

	wg.Wait()

	return nil
}

func downloadPartial(url string, bytesRangeFrom uint64, bytesRangeTo uint64) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download request. %w", err)
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", bytesRangeFrom, bytesRangeTo))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Download request failed. %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Download request failed. Status Code: %d", resp.StatusCode)
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
				return fmt.Errorf("Part %d: failed to write to file at offset %d: %w", partNo, currentWriteOffset, writeErr)
			}

			fmt.Printf("Part %d: Downloaded to file at offset %d\n", partNo, currentWriteOffset)

			currentWriteOffset += int64(bytesWritten)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			return fmt.Errorf("Part %d: failed to read response stream: %w", partNo, readErr)
		}
	}

	return nil
}
