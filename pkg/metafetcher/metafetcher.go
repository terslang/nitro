package metafetcher

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func FetchMetadata(url string) (MetaData, error) {
	var metadata MetaData
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
