package metafetcher

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
