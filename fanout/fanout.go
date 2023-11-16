// Package fanout sends a caller's request to all endpoints, either in series or parallel. When an
// endpoint's response is considered to be suited as end response (see package terminal) then
// that response is returned, and others are discarded.
package fanout

import (
	"log"
	"net/http"
	"sync"

	"github.com/KarelKubat/balancing-reverse-proxy/endpointresponse"
	"github.com/KarelKubat/balancing-reverse-proxy/endpoints"
	"github.com/KarelKubat/balancing-reverse-proxy/terminal"
)

type Fanout struct {
	ends   endpoints.Endpoints
	term   *terminal.Terminal
	runner func(w http.ResponseWriter, req *http.Request)
}

func New(inParallel bool, ends endpoints.Endpoints, term *terminal.Terminal) *Fanout {
	f := &Fanout{
		ends: ends,
		term: term,
	}
	if inParallel {
		f.runner = f.fanoutParallel
	} else {
		f.runner = f.fanoutSerial
	}
	return f
}

func (f *Fanout) Run(w http.ResponseWriter, req *http.Request) {
	log.Printf("serving request %v", req.URL)
	f.runner(w, req)
}

func (f *Fanout) fanoutParallel(w http.ResponseWriter, req *http.Request) {
	ch := make(chan *endpointresponse.EndpointResponse)
	var wg sync.WaitGroup
	for _, endp := range f.ends {
		wg.Add(1)
		go func(endp endpoints.Endpoint) {
			defer wg.Done()
			ch <- forwardToEndpoint(endp, w, req)
		}(endp)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	terminalSeen := false
	for er := range ch {
		if !terminalSeen && f.term.IsStop(er.Status) {
			log.Printf("endpoint returned a terminating status %v, discarding others", er.Status)
			w.Write(er.Body)
			terminalSeen = true
		}
	}
	if !terminalSeen {
		log.Printf("endpoints failed to return a valid answer, returning %v", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (f *Fanout) fanoutSerial(w http.ResponseWriter, req *http.Request) {
	var ers []*endpointresponse.EndpointResponse
	for _, endp := range f.ends {
		ers = append(ers, forwardToEndpoint(endp, w, req))
	}
	for i, er := range ers {
		if f.term.IsStop(er.Status) {
			log.Printf("endpoint %v returned a terminating status %v, discarding others", i, er.Status)
			w.Write(er.Body)
			return
		}
	}
	log.Printf("endpoints failed to return a valid answer, returning %v", http.StatusInternalServerError)
	w.WriteHeader(http.StatusInternalServerError)
}

func forwardToEndpoint(endp endpoints.Endpoint, w http.ResponseWriter, req *http.Request) *endpointresponse.EndpointResponse {
	er := endpointresponse.New(endp.URL, w.Header())
	endp.Proxy.ServeHTTP(er, req)
	return er
}
