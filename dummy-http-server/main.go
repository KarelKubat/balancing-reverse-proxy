package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	ms := rand.Intn(1000)
	sendErr := rand.Intn(100) < 25
	log.Printf("will serve %v after %vms, error:%v", req.URL, ms, sendErr)
	time.Sleep(time.Duration(ms) * time.Microsecond)
	if sendErr {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(fmt.Sprintf("Hello, world after %vms!", ms)))
}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8000", nil)
}
