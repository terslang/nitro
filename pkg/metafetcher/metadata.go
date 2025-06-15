package metafetcher

import (
	"github.com/dustin/go-humanize"

	"github.com/terslang/nitro/pkg/helpers"
)

type HttpMetaData struct {
	Url           string
	FileName      string
	ContentLength uint64
	AcceptRanges  bool
}

type FtpMetaData struct {
	Server        string
	Username      string
	Password      string
	FilePath      string
	FileName      string
	ContentLength uint64
}

func (m *HttpMetaData) LogMetaData() {
	helpers.Infof("File size: %s (%d bytes)\n", humanize.Bytes(m.ContentLength), m.ContentLength)
}

func (m *FtpMetaData) LogMetaData() {
	helpers.Infof("File size: %s (%d bytes)\n", humanize.Bytes(m.ContentLength), m.ContentLength)
}
