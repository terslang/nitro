package main

import (
	"fmt"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/terslang/nitro/pkg/helpers"
)

type ProgressBars struct {
	Container *mpb.Progress
	Bars      []*mpb.Bar
}

func MakeProgressBars(noOfPBars uint8, contentLength uint64) (*ProgressBars, error) {
	progress := mpb.New(mpb.WithAutoRefresh())

	partialContentSize, err := helpers.GetPartialContentSize(contentLength, noOfPBars)
	if err != nil {
		return nil, fmt.Errorf("error while making progress bars: %w", err)
	}

	bars := make([]*mpb.Bar, noOfPBars)

	for i := uint8(0); i < noOfPBars; i++ {
		rangeFromBytes, rangeToBytes := helpers.CalculateFromAndToBytes(contentLength, partialContentSize, i)
		barSize := int64(rangeToBytes - rangeFromBytes)
		bar := progress.New(
			barSize,
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("Part %d/%d", i+1, noOfPBars), decor.WC{W: 12, C: decor.DSyncWidth}),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{W: 5}),
				decor.CountersKibiByte("%.2f/%.2f", decor.WC{W: 20}),
			),
		)

		bars[i] = bar
	}

	return &ProgressBars{
		Container: progress,
		Bars:      bars,
	}, nil
}
