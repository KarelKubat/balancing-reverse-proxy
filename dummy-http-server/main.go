// Package main provides a dummy HTTP server for testing the balancing reverse proxy.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	flagAddress         = flag.String("address", ":8000", "address to bind to")
	flagStopAfter       = flag.Duration("stop-after", 0, "stop server once the duration expires, 0 is go on forever")
	flagDelayResponding = flag.Bool("delay-responding", true, "when true, fake a delay in processing")
	flagRandomErrors    = flag.Bool("random-errors", true, "simulate an error in (on average) 25% of cases")
)

func hello(w http.ResponseWriter, req *http.Request) {
	var sendErr bool
	var sleepTime time.Duration
	if *flagDelayResponding {
		sleepTime = time.Duration(rand.Intn(1000000)) * time.Microsecond
		sendErr = rand.Intn(100) < 25
		log.Printf("will serve %v after %v, error:%v", req.URL, sleepTime, sendErr)
		time.Sleep(sleepTime)
	}
	if sendErr {
		w.WriteHeader(http.StatusInternalServerError)
	}
	var msg string
	if sleepTime == 0 {
		msg = fmt.Sprintf("Hello, world from %v%v\n", *flagAddress, req.URL)
	} else {
		msg = fmt.Sprintf("Hello, world from %v%v after %v\n", *flagAddress, req.URL, sleepTime)
	}
	w.Write([]byte(msg))
}

func main() {
	flag.Parse()
	if *flagStopAfter > 0 {
		go func() {
			time.Sleep(*flagStopAfter)
			fmt.Printf("stopping server on %q after %v\n", *flagAddress, *flagStopAfter)
			os.Exit(0)
		}()
	}
	http.HandleFunc("/", hello)
	http.ListenAndServe(*flagAddress, nil)
}
