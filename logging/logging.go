// Package logging provides a structured way to initialize the default log package.
package logging

import (
	"io"
	"log"
	"os"
)

type Opts struct {
	LogPrefix string
	LogFile   string
	LogDate   bool
	LogTime   bool
	LogMsec   bool
	LogUTC    bool
}

func Setup(opts *Opts) error {
	if opts.LogPrefix != "" {
		log.SetPrefix(opts.LogPrefix + " ")
	}

	log.SetFlags(loggingFlags(opts))

	var wr io.Writer
	switch opts.LogFile {
	case "", "stdout":
		wr = os.Stdout
	case "stderr":
		wr = os.Stderr
	default:
		var err error
		wr, err = os.OpenFile(opts.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
	}
	log.SetOutput(wr)

	return nil
}

// loggingFlags is a (testable) helper to determine the flags for `log.SetFlags`.
func loggingFlags(opts *Opts) int {
	var logFlags int
	if opts.LogDate {
		logFlags |= log.Ldate
	}
	if opts.LogTime {
		logFlags |= log.Ltime
	}
	if opts.LogMsec {
		logFlags |= log.Ltime
		logFlags |= log.Lmicroseconds
	}
	if opts.LogUTC {
		logFlags |= log.LUTC
	}
	return logFlags
}
