package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	sleepTime := time.Duration(rand.Intn(1000000)) * time.Microsecond
	sendErr := rand.Intn(100) < 25
	log.Printf("will serve %v after %v, error:%v", req.URL, sleepTime, sendErr)
	time.Sleep(sleepTime)
	if sendErr {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(fmt.Sprintf("Hello, world after %v!", sleepTime)))
}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8000", nil)
}
