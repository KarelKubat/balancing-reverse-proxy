// Package fanout sends a caller's request to all endpoints, either in series or parallel. When an
// endpoint's response is considered to be suited as end response (see package terminal) then
// that response is returned, and others are discarded.
package fanout

import (
	"log"
	"net/http"
	"time"

	"github.com/KarelKubat/balancing-reverse-proxy/endpointresponse"
	"github.com/KarelKubat/balancing-reverse-proxy/endpoints"
	"github.com/KarelKubat/balancing-reverse-proxy/terminal"
	"github.com/KarelKubat/puddle"
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
	elapsed := time.Since(start)
	log.Printf("request %v served in %v by endpoint %v", req.URL, elapsed, i)
}

type puddleResponse struct {
	resp       *endpointresponse.EndpointResponse
	endpointNr int
}

func puddleWorker(a puddle.Args) any {
	// These MUST be the right types in the right order, just as the below p.Work() sends them
	nr := a[0].(int)
	endp := a[1].(endpoints.Endpoint)
	w := a[2].(http.ResponseWriter)
	req := a[3].(*http.Request)
	return puddleResponse{
		resp:       forwardToEndpoint(endp, w, req),
		endpointNr: nr,
	}
}

func (f *Fanout) fanoutParallel(w http.ResponseWriter, req *http.Request) int {
	p := puddle.New()
	for i, endp := range f.ends {
		// These MUST BE the right types in the right order so that puddleWorker
		// can decode them: ------------ >  >>>>  >  >>>
		p.Work(puddleWorker, puddle.Args{i, endp, w, req})
	}
	endpointNr := -1 // assume all endpoints failed
	for v := range p.Out() {
		er := v.(puddleResponse)
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
