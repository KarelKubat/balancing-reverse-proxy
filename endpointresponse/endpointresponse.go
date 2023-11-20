// Package endpointresponse represents what an endpoint replies. It implements the interface
// http.ResponseWriter (methods: Header(), Write(), WriteHeader()). An EndpointResponse can therefore
// be used in a reverse proxy's ServeHTTP().
package endpointresponse

import (
	"net/http"
)

type EndpointResponse struct {
	URL    string
	Body   []byte
	Status int
	header http.Header
}

func New(URL string, h http.Header) *EndpointResponse {
	return &EndpointResponse{
		URL:    URL,
		Body:   []byte{},
		header: h.Clone(), // A header is a map. We need a clone to avoid map read/write race conditions.
		Status: 0,
	}
}

func (e *EndpointResponse) Header() http.Header {
	return e.header
}

func (e *EndpointResponse) Write(b []byte) (int, error) {
	if e.Status == 0 {
		// log.Printf("endpoint %q sent a first write, assuming status OK", e.URL)
		e.Status = http.StatusOK
	}
	// log.Printf("endpoint %q sent %v bytes (status: %v)", e.URL, len(b), e.Status)
	e.Body = append(e.Body, b...)
	return len(e.Body), nil
}

func (e *EndpointResponse) WriteHeader(c int) {
	e.Status = c
}
