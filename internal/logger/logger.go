package logger

import (
	"log"
	"os"
)

func Init() *log.Logger {
	return log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}
