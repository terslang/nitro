package logger

import (
	"log"
	"os"
)

var Verbose = false

var debugLog = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)
var infoLog = log.New(os.Stdout, "", 0)

func Debug(format string, v ...any) {
	if Verbose {
		debugLog.Printf(format, v...)
	}
}

func Infof(format string, v ...any) {
	infoLog.Printf(format, v...)
}

func Infoln(v ...any) {
	infoLog.Println(v...)
}
