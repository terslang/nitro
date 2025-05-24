package helpers

import "fmt"

func GetPartialContentSize(contentLength uint64, parallel uint8) (uint64, error) {
	if parallel == 0 {
		return 0, fmt.Errorf("Parallel: 0, Division by zero")
	}

	return (contentLength + uint64(parallel) - 1) / uint64(parallel), nil
}

func CalculateFromAndToBytes(contentLength uint64, partialContentSize uint64, partNo uint8) (uint64, uint64) {
	rangeFromBytes := partialContentSize * uint64(partNo)
	rangeToBytes := rangeFromBytes + partialContentSize - 1

	rangeToBytes = min(rangeToBytes, contentLength)

	return rangeFromBytes, rangeToBytes
}
