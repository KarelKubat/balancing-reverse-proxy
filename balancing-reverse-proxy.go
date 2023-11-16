// Package main is the entry point for the balancing reverse HTTP proxy.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/KarelKubat/balancing-reverse-proxy/endpoints"
	"github.com/KarelKubat/balancing-reverse-proxy/fanout"
	"github.com/KarelKubat/balancing-reverse-proxy/terminal"
	"github.com/KarelKubat/flagnames"
)

var (
	flagEndpoints         = flag.String("endpoints", "", "comma-separated list of endpoints, ex. 'https://one.com,https://two.com'")
	flagAddress           = flag.String("address", ":8080", "address to bind this proxy")
	flagTerminalResponses = flag.String("terminal-responses", "100,200,300,400", "HTTP statuses that are considered terminal (i.e., that endpoint's response is taken)")
	flagParallelWorkers   = flag.Bool("parallel-workers", true, "when true, workers for endpoints start in parallel, else in sequence")
	flagLogPrefix         = flag.String("log-prefix", "balancing-reverse-proxy", "prefix for log statements")
)

const (
	usage = `
Minimum usage: balancing-reverse-proxy --endpoints=URLs
Flags may be abbreviated, e.g. "-e" for "-endpoints". Available flags are:

`
)

func main() {
	// Parse commandline.
	flagnames.Patch()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if *flagEndpoints == "" || flag.NArg() != 0 || *flagTerminalResponses == "" {
		flag.Usage()
	}

	// Logging.
	if *flagLogPrefix != "" {
		log.SetPrefix(*flagLogPrefix + " ")
	}

	// Which of https://developer.mozilla.org/en-US/docs/Web/HTTP/Status indicate that an endpoint's response should be given to the caller?
	term, err := terminal.New(*flagTerminalResponses)
	check(err)

	// Which endpoints do we have?
	ends, err := endpoints.New(*flagEndpoints)
	check(err)

	// Instantiate the fanout engine.
	f := fanout.New(*flagParallelWorkers, ends, term)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		f.Run(w, req)
	})

	// Serve.
	log.Printf("starting on: %q", *flagAddress)
	if err := http.ListenAndServe(*flagAddress, nil); err != nil {
		log.Fatal("cannot bind to address:", err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
