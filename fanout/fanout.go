// Package fanout sends a caller's request to all endpoints, either in series or parallel. When an
// endpoint's response is considered to be suited as end response (see package terminal) then
// that response is returned, and others are discarded.
package fanout

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KarelKubat/balancing-reverse-proxy/endpointresponse"
	"github.com/KarelKubat/balancing-reverse-proxy/endpoints"
	"github.com/KarelKubat/balancing-reverse-proxy/terminal"
)

type Fanout struct {
	ends   endpoints.Endpoints
	term   *terminal.Terminal
	runner func(w http.ResponseWriter, req *http.Request) int
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
	start := time.Now()
	i := f.runner(w, req)
	elapsed := time.Now().Sub(start)
	log.Printf("request %v served in %v by endpoint %v", req.URL, elapsed, i)
}

type parallelResponse struct {
	resp       *endpointresponse.EndpointResponse
	endpointNr int
}

func (f *Fanout) fanoutParallel(w http.ResponseWriter, req *http.Request) int {
	ch := make(chan *parallelResponse)
	var wg sync.WaitGroup
	for i, endp := range f.ends {
		wg.Add(1)
		go func(endp endpoints.Endpoint, endpointNr int) {
			defer wg.Done()
			ch <- &parallelResponse{
				resp:       forwardToEndpoint(endp, w, req),
				endpointNr: endpointNr,
			}
		}(endp, i)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	endpointNr := -1 // assume all endpoints failed
	for er := range ch {
		if endpointNr == -1 && f.term.IsStop(er.resp.Status) {
			sendResponse(w, er.resp)
			endpointNr = er.endpointNr
		}
	}
	if endpointNr == -1 {
		log.Printf("endpoints failed to return a valid answer, returning %v", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
	}
	return endpointNr
}

func (f *Fanout) fanoutSerial(w http.ResponseWriter, req *http.Request) int {
	var ers []*endpointresponse.EndpointResponse
	for _, endp := range f.ends {
		ers = append(ers, forwardToEndpoint(endp, w, req))
	}
	for i, er := range ers {
		if f.term.IsStop(er.Status) {
			sendResponse(w, er)
			return i
		}
	}
	log.Printf("endpoints failed to return a valid answer, returning %v", http.StatusInternalServerError)
	w.WriteHeader(http.StatusInternalServerError)
	return -1
}

func forwardToEndpoint(endp endpoints.Endpoint, w http.ResponseWriter, req *http.Request) *endpointresponse.EndpointResponse {
	er := endpointresponse.New(endp.URL, w.Header())
	endp.Proxy.ServeHTTP(er, req)
	return er
}

func sendResponse(w http.ResponseWriter, er *endpointresponse.EndpointResponse) {
	// log.Printf("endpoint returned a terminating status %v, discarding others", er.Status)
	if er.Status != http.StatusOK {
		http.Error(w, "", er.Status)
	}
	w.Write(er.Body)
}
