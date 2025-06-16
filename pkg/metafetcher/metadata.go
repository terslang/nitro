package metafetcher

import (
	"github.com/atuleu/go-humanize"

	"github.com/terslang/nitro/pkg/logger"
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
	logger.Infof("File size: %s (%d bytes)\n", humanize.ByteSize(m.ContentLength), m.ContentLength)
}

func (m *FtpMetaData) LogMetaData() {
	logger.Infof("File size: %s (%d bytes)\n", humanize.ByteSize(m.ContentLength), m.ContentLength)
}
