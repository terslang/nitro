package helpers

import (
	"fmt"
	"log"
	"os"
)

var Verbose = false

func ceilDiv(a uint64, b uint64) (uint64, error) {
	if b == 0 {
		return 0, fmt.Errorf("Denominator: 0, Division by zero")
	}

	return (a + b - 1) / b, nil
}

func GetPartialContentSize(contentLength uint64, parallel uint8) (uint64, error) {
	partialContentSize, err := ceilDiv(contentLength, uint64(parallel))
	if err != nil {
		return 0, fmt.Errorf("Error calculating partial content size: %w", err)
	}

	return partialContentSize, nil
}

func CalculateFromAndToBytes(contentLength uint64, partialContentSize uint64, partNo uint8) (uint64, uint64) {
	rangeFromBytes := partialContentSize * uint64(partNo)
	rangeToBytes := rangeFromBytes + partialContentSize - 1

	rangeToBytes = min(rangeToBytes, contentLength-1)

	return rangeFromBytes, rangeToBytes
}

var debugLog = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)

func Debug(format string, v ...any) {
	if Verbose {
		debugLog.Printf(format, v...)
	}
}

var infoLog = log.New(os.Stdout, "", 0)

func Infof(format string, v ...any) {
	infoLog.Printf(format, v...)
}

func Infoln(v ...any) {
	infoLog.Println(v...)
}
