package options

const DefaultFileName string = "output-file"

type NitroOptions struct {
	Url            string
	Parallel       uint8
	OutputFileName string
}
