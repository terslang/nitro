package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/jlaffaye/ftp"
	"github.com/schollz/progressbar/v3"
	"github.com/terslang/nitro/pkg/helpers"
	"github.com/terslang/nitro/pkg/metafetcher"
	"github.com/terslang/nitro/pkg/options"
)

const downloadBufferSize int = 1024 * 1024 // 1mb buffer

func DownloadHttp(metadata metafetcher.HttpMetaData, opts options.NitroOptions) error {
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
		return fmt.Errorf("Failed to create file: %w", err)
	}
	defer outFile.Close()

	if metadata.ContentLength > 0 {
		if err := outFile.Truncate(int64(metadata.ContentLength)); err != nil {
			return fmt.Errorf("Failed to preallocate file space: %w", err)
		}
	}

	bar := progressbar.NewOptions64(
		int64(metadata.ContentLength),
		progressbar.OptionSetDescription("Downloading..."),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
	)

	var wg sync.WaitGroup
	errorChannel := make(chan error, parallel)

	partialContentSize, err := helpers.GetPartialContentSize(metadata.ContentLength, parallel)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	httpClient := GetHttpClient(parallel)

	for i := uint8(0); i < parallel; i++ {
		wg.Add(1)
		go func(partNo uint8) {
			defer wg.Done()

			fmt.Println("Starting partial download", partNo)

			rangeFromBytes, rangeToBytes := helpers.CalculateFromAndToBytes(metadata.ContentLength, partialContentSize, partNo)

			partialContentsReader, err := downloadPartialHttp(metadata.Url, httpClient, rangeFromBytes, rangeToBytes)
			if err != nil {
				err = fmt.Errorf("Part %d (range %d-%d): failed to initiate download: %w", partNo, rangeFromBytes, rangeToBytes, err)
				errorChannel <- err
				return
			}

			err = downloadAndWriteToFileTilEof(partialContentsReader, outFile, int64(rangeFromBytes), partNo, rangeToBytes-rangeFromBytes+1, bar)
			if err != nil {
				errorChannel <- fmt.Errorf("Part %d (range %d-%d): failed to download: %w", partNo, rangeFromBytes, rangeToBytes, err)
				return
			}

			fmt.Printf("Part %d is done\n", partNo)
		}(i)
	}

	wg.Wait()
	close(errorChannel)

	var firstError error
	for errFromGoRoutine := range errorChannel {
		if errFromGoRoutine != nil && firstError == nil {
			firstError = errFromGoRoutine
		}
	}

	if firstError != nil {
		return fmt.Errorf("Download of %s failed: %w", fileName, firstError)
	}

	return nil
}

func downloadPartialHttp(url string, http_client *http.Client, bytesRangeFrom uint64, bytesRangeTo uint64) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download request. %w", err)
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", bytesRangeFrom, bytesRangeTo))

	resp, err := http_client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Download request failed. %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Download request failed. Status Code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func DownloadFtp(metadata metafetcher.FtpMetaData, opts options.NitroOptions) error {
	var fileName string
	if opts.OutputFileName == options.DefaultFileName {
		fileName = metadata.FileName
	} else {
		fileName = opts.OutputFileName
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Failed to create file. %w", err)
	}
	defer outFile.Close()

	var wg sync.WaitGroup
	errorChannel := make(chan error, opts.Parallel)

	partialContentSize, err := helpers.GetPartialContentSize(metadata.ContentLength, opts.Parallel)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for i := uint8(0); i < opts.Parallel; i++ {
		wg.Add(1)
		go func(partNo uint8) {
			defer wg.Done()

			fmt.Println("Starting partial download", partNo)

			rangeFromBytes, _ := helpers.CalculateFromAndToBytes(metadata.ContentLength, partialContentSize, partNo)
			conn, partialContentsReader, err := downloadPartialFtp(metadata, rangeFromBytes)

			if err != nil {
				err = fmt.Errorf("Part %d (offset %d): failed to initiate download: %w", partNo, rangeFromBytes, err)
				errorChannel <- err
				return
			}

			defer conn.Quit()

			err = downloadAndWriteToFileTilSize(partialContentsReader, outFile, int64(rangeFromBytes), partNo, partialContentSize)
			if err != nil {
				err = fmt.Errorf("Part %d (offset %d): failed to download: %w", partNo, rangeFromBytes, err)
				errorChannel <- err
				return
			}

			fmt.Printf("%d Partial download is done\n", partNo)
		}(i)
	}

	wg.Wait()

	close(errorChannel)

	var firstError error
	for errFromGoRoutine := range errorChannel {
		if errFromGoRoutine != nil {
			if firstError == nil {
				firstError = errFromGoRoutine
			}
		}
	}

	if firstError != nil {
		return fmt.Errorf("Download of %s failed: %w", fileName, firstError)
	}

	return nil
}

func downloadPartialFtp(metadata metafetcher.FtpMetaData, rangeFromBytes uint64) (*ftp.ServerConn, io.ReadCloser, error) {
	conn, err := ftp.Dial(metadata.Server)
	if err != nil {
		return nil, nil, fmt.Errorf("Error connecting to the server: %w", err)
	}

	if metadata.Username != "" {
		err = conn.Login(metadata.Username, metadata.Password)
		if err != nil {
			return nil, nil, fmt.Errorf("Error logging in: %w", err)
		}
	}

	resp, err := conn.RetrFrom(metadata.FilePath, rangeFromBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Retr request failed: %w", err)
	}

	return conn, resp, nil
}

func downloadAndWriteToFileTilEof(reader io.ReadCloser, file *os.File, startOffset int64, partNo uint8, expectedSize uint64, bar *progressbar.ProgressBar) error {
	defer reader.Close()

	buffer := make([]byte, downloadBufferSize)
	currentWriteOffset := startOffset

	for {
		bytesRead, readErr := reader.Read(buffer)
		if bytesRead > 0 {
			bytesWritten, writeErr := file.WriteAt(buffer[:bytesRead], currentWriteOffset)
			if writeErr != nil {
				return fmt.Errorf("Part %d: failed to write to file at offset %d: %w", partNo, currentWriteOffset, writeErr)
			}

			currentWriteOffset += int64(bytesWritten)
			bar.Add(bytesWritten)
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

func downloadAndWriteToFileTilSize(reader io.ReadCloser, file *os.File, startOffset int64, partNo uint8, partialContentSize uint64) error {
	defer reader.Close()

	buffer := make([]byte, downloadBufferSize)
	currentWriteOffset := startOffset
	totalBytesWritten := uint64(0)

	for {
		bytesRead, readErr := reader.Read(buffer)

		if bytesRead > 0 {
			bytesToWrite := min(bytesRead, int(partialContentSize-totalBytesWritten))
			bytesWritten, writeErr := file.WriteAt(buffer[:bytesToWrite], currentWriteOffset)
			if writeErr != nil {
				return fmt.Errorf("Part %d: failed to write to file at offset %d: %w", partNo, currentWriteOffset, writeErr)
			}

			fmt.Printf("Part %d: Downloaded to file at offset %d\n", partNo, currentWriteOffset)

			currentWriteOffset += int64(bytesWritten)
			totalBytesWritten += uint64(bytesWritten)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			return fmt.Errorf("Part %d: failed to read response stream: %w", partNo, readErr)
		}

		if totalBytesWritten >= partialContentSize {
			break
		}
	}

	return nil
}
