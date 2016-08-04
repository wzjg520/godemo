package main

import (
	hd "batch/httphandler"
	"log"
	"net/http"
	"runtime/debug"
)

const (
	PORT = ":3210"
)

//var imagesChan chan string = make(chan string, 30)

func wrapHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, " ", r.Proto, " ", r.URL.Scheme, " ", r.Host, " ", r.URL.Path)
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Println("WARN: panic in %v - %v", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(w, r)

	}
}

func main() {

	httpHandler := hd.NewHttpHandler()
	httpHandler.Init(10, 30)

	mux := http.NewServeMux()
	mux.HandleFunc("/batch/save_images", wrapHandler(httpHandler.SaveImages))
	err := http.ListenAndServe(PORT, mux)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("start listen ", PORT)
}
