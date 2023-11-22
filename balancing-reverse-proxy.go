// Package main is the entry point for the balancing reverse HTTP proxy.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/KarelKubat/balancing-reverse-proxy/endpoints"
	"github.com/KarelKubat/balancing-reverse-proxy/fanout"
	"github.com/KarelKubat/balancing-reverse-proxy/logging"
	"github.com/KarelKubat/balancing-reverse-proxy/terminal"
	"github.com/KarelKubat/flagnames"
)

var (
	flagEndpoints         = flag.String("endpoints", "", "comma-separated list of endpoints, ex. 'https://one.com,https://two.com'")
	flagAddress           = flag.String("address", ":8080", "address to bind this proxy")
	flagTerminalResponses = flag.String("terminal-responses", "100,200,300,400", "HTTP statuses that are considered terminal (i.e., that endpoint's response is taken)")
	flagFanout            = flag.Bool("fanout", false, "when true, workers for endpoints start in parallel, else in sequence")
	flagLogPrefix         = flag.String("log-prefix", "balancing-reverse-proxy", "prefix for log statements")
	flagLogFile           = flag.String("log-file", "stdout", "log output, a true file to append, or `stdout` or `stderr`")
	flagLogDate           = flag.Bool("log-date", true, "when true, emit the date when logging")
	flagLogTime           = flag.Bool("log-time", true, "when true, emit the time when logging")
	flagLogMsec           = flag.Bool("log-msec", false, "when true, emit the microseconds when logging (forces -log-time)")
	flagLogUTC            = flag.Bool("log-utc", false, "when true, log date and/or time in UTC rather than localtime")
	flagStopAfter         = flag.Duration("stop-after", 0, "stop once the duration expires, 0 is go on forever (for testing)")
)

const (
	usage = `
Minimum usage: balancing-reverse-proxy --endpoints=URLs
Flags may be abbreviated, e.g. "-e" for "-endpoints", but abbreviations must be unique ("-log" could mean many things, "-log-f" is okay).
Available flags are:

`
)

func main() {
	// Parse commandline.
	flagnames.Patch()
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if *flagEndpoints == "" || flag.NArg() != 0 || *flagTerminalResponses == "" {
		flag.Usage()
	}
	check(logging.Setup(&logging.Opts{
		LogPrefix: *flagLogPrefix,
		LogFile:   *flagLogFile,
		LogDate:   *flagLogDate,
		LogTime:   *flagLogTime,
		LogMsec:   *flagLogMsec,
		LogUTC:    *flagLogUTC,
	}))

	// Which of https://developer.mozilla.org/en-US/docs/Web/HTTP/Status indicate that an endpoint's response should be given to the caller?
	term, err := terminal.New(*flagTerminalResponses)
	check(err)

	// Which endpoints do we have?
	ends, err := endpoints.New(*flagEndpoints)
	check(err)

	// Instantiate the fanout engine.
	f := fanout.New(*flagFanout, ends, term)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		f.Run(w, req)
	})

	// If a stop is requested..
	if *flagStopAfter > 0 {
		go func() {
			time.Sleep(*flagStopAfter)
			fmt.Printf("stopping balancing-reverse-proxy after %v\n", *flagStopAfter)
			os.Exit(0)
		}()
	}

	// Serve.
	log.Printf("starting on: %q", *flagAddress)
	if err := http.ListenAndServe(*flagAddress, nil); err != nil {
		log.Fatal("cannot bind to address:", err)
	}
}

// check is a helper to abort main() when an error happens.
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
