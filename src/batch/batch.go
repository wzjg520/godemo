package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

// port 端口
const (
	PORT = ":3210"
)

var cacheDir = flag.String("d", "/tmp", "临时文件存放目录")
var scriptPath = flag.String("f", "/home/john/save_img.php", "执行脚本路径")

func wrapHandler(pool downloaderPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, " ", r.Proto, " ", r.URL.Scheme, " ", r.Host, " ", r.URL.Path)
		if r.Method == "POST" {
			var urls []string
			err := parseBodySlice(r, &urls)
			if err != nil {
				log.Println(err)
				io.WriteString(w, err.Error())
				return
			}

			start := time.Now().Unix()
			urlsChan := make(chan map[string]string, 1)
			defer func() {
				close(urlsChan)
			}()
			for _, v := range urls {
				go startDownload(pool, urlsChan, v)
			}

			var resData []map[string]string
			for _ = range urls {
				select {
				case imgURL := <-urlsChan:
					resData = append(resData, imgURL)
				}
			}

			end := time.Now().Unix()
			sub := end - start

			log.Printf("costTime: %d", sub)
			jsonData, _ := json.Marshal(resData)
			io.WriteString(w, string(jsonData))

		} else {
			io.WriteString(w, "GET method is not support!")
		}
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Println("WARN: panic in %v - %v", pool, e)
				log.Println(string(debug.Stack()))
			}
		}()
	}
}

// 开始批量下载
func startDownload(pool downloaderPool, urlsChan chan map[string]string, url string) {
	dpl, err := pool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("The pool can not return a entity")
		log.Fatalf(errMsg)
	}
	defer func() {
		err := pool.Return(dpl)
		if err != nil {
			log.Print(err)
		}

		if p := recover(); p != nil {
			log.Println(p)
		}
	}()

	urlsChan <- dpl.SaveImages(url)

}

func genDownloader() imgDownloader {
	return NewDownloader()
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

	dlPool, err := NewDownloaderPool(50, genDownloader)
	// 监控用量
	go func() {
		for {
			log.Printf("watching: total %d, used %d, goroutine %d,", dlPool.Total(), dlPool.Used(), runtime.NumGoroutine())
			time.Sleep(5 * time.Second)
		}
	}()
	if err != nil {
		log.Fatalf("init pool error!")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/batch/save_images", wrapHandler(dlPool))
	log.Println("start listen localhost", PORT)
	log.Printf("cache file will save in %s", *cacheDir)
	log.Printf("script file %s", *scriptPath)
	err = http.ListenAndServe(PORT, mux)
	if err != nil {
		log.Fatal(err.Error())
	}

}
