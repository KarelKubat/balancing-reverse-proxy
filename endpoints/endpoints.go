// Package endpoints represents the configured end points. Each of these receives an instance
// of httputil.NewSingleHostReverseProxy as a dedicated proxy.
package endpoints

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"net/http/httputil"
)

type Endpoint struct {
	URL   string
	Proxy *httputil.ReverseProxy
}

type Endpoints []Endpoint

func New(set string) (Endpoints, error) {
	var e Endpoints
	for i, endp := range strings.Split(set, ",") {
		if endp == "" {
			return e, errors.New("empty endpoint")
		}
		log.Printf("configuring endpoint %v: %q", i, endp)
		target, err := url.Parse(endp)
		if err != nil {
			return e, fmt.Errorf("bad endpoint %q: %v", endp, err)
		}
		e = append(e, Endpoint{
			URL:   endp,
			Proxy: httputil.NewSingleHostReverseProxy(target),
		})
	}
	return e, nil
}
