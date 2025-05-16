package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/akamensky/argparse"
)

func parseCommandLineInputs(args []string) (nitroOptions, error) {
	parser := argparse.NewParser("nitro", "Download Accelerator")

	urlArg := parser.StringPositional(nil)
	parallelArg := parser.Int("p", "parallel", &argparse.Options{
		Required: false,
		Validate: func(args []string) error {
			arg, err := strconv.Atoi(args[0])

			if err != nil {
				return err
			}

			if arg < 0 || arg > math.MaxUint8 {
				return fmt.Errorf("Invalid parallel value")
			}

			return nil
		},
		Help: "Number of parallel downloads",
	})

	err := parser.Parse(args)

	if err != nil {
		fmt.Println("failed to parse the command")
		return nitroOptions{}, err
	}

	var url string
	if urlArg != nil {
		url = *urlArg
	} else {
		url = ""
	}

	var parallel uint8
	if parallelArg != nil {
		parallel = uint8(*parallelArg)
	} else {
		parallel = 0
	}

	return nitroOptions{
		Url:      url,
		Parallel: parallel,
	}, nil
}
