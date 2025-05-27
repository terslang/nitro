package downloader

import (
	"net/http"
	"time"
)

func GetHttpClient(parallel uint8) *http.Client {
	if parallel > 1 {
		return &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:      int(parallel),
				IdleConnTimeout:   90 * time.Second,
				DisableKeepAlives: false,
			},
		}
	}
	return &http.Client{}
}
