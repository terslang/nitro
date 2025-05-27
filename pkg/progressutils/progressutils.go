package progressutils

import (
	"fmt"
	"strings"
	"sync"
)

type SegmentedProgressBar struct {
	Total           uint64
	SegmentCount    int
	SegmentTargets  []uint64 // expected bytes per segment
	SegmentProgress []uint64 // bytes downloaded so far for each segment
	mu              sync.Mutex
}

func NewSegmentedProgressBar(total uint64, segmentTargets []uint64, filename string) *SegmentedProgressBar {
	segCount := len(segmentTargets)
	return &SegmentedProgressBar{
		Total:           total,
		SegmentCount:    segCount,
		SegmentTargets:  segmentTargets,
		SegmentProgress: make([]uint64, segCount),
	}
}

func (pb *SegmentedProgressBar) UpdateSegment(segment int, bytesAdded uint64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if segment < 0 || segment >= pb.SegmentCount {
		// Out-of-bound segment index; ignore or handle error as needed.
		return
	}

	// Calculate how many bytes can actually be added to this segment.
	remaining := pb.SegmentTargets[segment] - pb.SegmentProgress[segment]
	if bytesAdded > remaining {
		bytesAdded = remaining
	}
	pb.SegmentProgress[segment] += bytesAdded
}

func (pb *SegmentedProgressBar) Render() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	// Each segment renders as a mini-bar of fixed width.
	barWidth := 10
	segmentsStr := make([]string, pb.SegmentCount)
	for i := 0; i < pb.SegmentCount; i++ {
		ratio := float64(pb.SegmentProgress[i]) / float64(pb.SegmentTargets[i])
		filled := int(ratio * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		segBar := "[" + strings.Repeat("█", filled) + strings.Repeat(" ", barWidth-filled) + "]"
		segmentsStr[i] = segBar
	}
	// Calculate overall progress percentage.
	var sum uint64 = 0
	for i := 0; i < pb.SegmentCount; i++ {
		sum += pb.SegmentProgress[i]
	}
	overall := float64(sum) / float64(pb.Total) * 100
	return fmt.Sprintf("%s  %.2f%%", strings.Join(segmentsStr, " "), overall)
}

func (pb *SegmentedProgressBar) Finish() {
	fmt.Println()
	fmt.Println("Download completed successfully!")
}
