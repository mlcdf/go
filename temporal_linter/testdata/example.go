package main

import (
	"log"
)

type Logger struct{}

func (l Logger) Debug(format string, args ...any) {
	log.Printf(format, args...)
}

func main() {
	logger := &Logger{}
	logger.Debug("dddd")
	logger.Debug("dd%d")      // want "logger.Debug call has possible logf-style formatting directive %d"
	logger.Debug("dd%s")      // want "logger.Debug call has possible logf-style formatting directive %s"
	logger.Debug("dd", "key") // want "invalid log usage"
	logger.Debug("dd", "key", "value")
}
