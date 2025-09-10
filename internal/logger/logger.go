package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func Init() error {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
