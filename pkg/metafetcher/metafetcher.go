package metafetcher

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func FetchMetadata(url string) MetaData {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println("Error while fetching metadata")
	}

	var filename string
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		if strings.HasPrefix(contentDisposition, "attachment;") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				filename = strings.Trim(strings.Trim(parts[1], "\""), " ")
			}
		}
	}

	var contentLength uint64
	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader != "" {
		contentLength, _ = strconv.ParseUint(contentLengthHeader, 10, 64)
	}

	return MetaData{
		Url:           url,
		FileName:      filename,
		ContentLength: contentLength,
	}
}
