package main

import (
	hd "batch/httphandler"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

const (
	PORT = ":3210"
)

var cacheDir *string = flag.String("d", "/tmp", "临时文件存放目录")
var scriptPath *string = flag.String("f", "/home/john/save_img.php", "执行脚本路径")

func wrapHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, " ", r.Proto, " ", r.URL.Scheme, " ", r.Host, " ", r.URL.Path)
		hdl := hd.NewHttpHandler(*cacheDir, *scriptPath)
		hdl.SaveImages(w, r)
		defer func(hdl *hd.HttpHandler) {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Println("WARN: panic in %v - %v", hdl.SaveImages, e)
				log.Println(string(debug.Stack()))
			}
		}(hdl)


	}
}

func main() {
	flag.Parse()

	_, err := os.Stat(*cacheDir)

	if err != nil {
		log.Fatalf("dir %s not existes!", *cacheDir)
	}

	_, err = os.Stat(*scriptPath)

	if err != nil {
		log.Fatalf("script %s not existes!", *scriptPath)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/batch/save_images", wrapHandler())
	log.Println("start listen localhost", PORT)
	log.Printf("cache file will save in %s", *cacheDir)
	log.Printf("script file %s", *scriptPath)
	err = http.ListenAndServe(PORT, mux)
	if err != nil {
		log.Fatal(err.Error())
	}

}
