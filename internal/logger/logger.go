package logger

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger

func Create(path string) error {
	if len(path) == 0 {
		Logger = log.New(io.Discard, "", 0)
		return nil
	}

	logFile, err := os.OpenFile(path+"/ascii-graphics.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
