package metafetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jlaffaye/ftp"

	"github.com/terslang/nitro/pkg/helpers"
)

func FetchMetadataHttp(url string) (HttpMetaData, error) {
	helpers.Infoln("Fetching Metadata...")
	var metadata HttpMetaData
	resp, err := http.Head(url)
	if err != nil {
		return metadata, fmt.Errorf("Error while fetching metadata. %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return metadata, fmt.Errorf("Metadata request failed. Response status code: %d", resp.StatusCode)
	}

	metadata.Url = url

	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		if strings.HasPrefix(contentDisposition, "attachment;") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				metadata.FileName = strings.Trim(strings.Trim(parts[1], "\""), " ")
			}
		}
	} else {
		urlParts := strings.Split(url, "/")
		metadata.FileName = strings.Trim(urlParts[len(urlParts)-1], " ")
	}

	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader != "" {
		metadata.ContentLength, err = strconv.ParseUint(contentLengthHeader, 10, 64)
		if err != nil {
			return metadata, fmt.Errorf("Failed to read content length from the metadata. %w", err)
		}
	} else {
		metadata.ContentLength = 0
	}

	acceptRangesHeader := resp.Header.Get("Accept-Ranges")
	if acceptRangesHeader == "bytes" {
		metadata.AcceptRanges = true
	} else {
		metadata.AcceptRanges = false
	}

	return metadata, nil
}

func FetchMetadataFtp(rawUrl string) (FtpMetaData, error) {
	var metadata FtpMetaData

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return metadata, fmt.Errorf("Failed to parse the URL: %w", err)
	}

	metadata.Server = parsedUrl.Host

	if !strings.Contains(metadata.Server, ":") {
		metadata.Server += ":21"
	}

	metadata.Username = parsedUrl.User.Username()
	metadata.Password, _ = parsedUrl.User.Password()
	metadata.FilePath = parsedUrl.Path
	pathParts := strings.Split(metadata.FilePath, "/")
	metadata.FileName = pathParts[len(pathParts)-1]

	helpers.Infoln("Fetching Metadata...")

	conn, err := ftp.Dial(metadata.Server)
	if err != nil {
		return metadata, fmt.Errorf("Error connecting to the server: %w", err)
	}
	defer conn.Quit()

	if metadata.Username != "" {
		err = conn.Login(metadata.Username, metadata.Password)
		if err != nil {
			return metadata, fmt.Errorf("Error logging in: %w", err)
		}
	}

	fileSize, err := conn.FileSize(metadata.FilePath)
	if err != nil {
		return metadata, fmt.Errorf("Error while retreiving file size: %w", err)
	}

	metadata.ContentLength = uint64(fileSize)

	return metadata, nil
}
