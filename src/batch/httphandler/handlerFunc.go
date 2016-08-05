package httphandler

import (
	"net/http"
	//"encoding/json"
	"io"
	"log"
	"encoding/json"
	"runtime"
	"time"
)

type HttpHandler struct {
	RequestChan chan string
	ResultChan  chan map[string]string
}

// 生成资源对象
func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

// 初始化通道
func (hd *HttpHandler) InitChan(requestChanLength int, resultChanLength int) {
	log.Printf("init resultChan length %d", resultChanLength)
	log.Printf("init a requestChan length %d", requestChanLength)
	hd.RequestChan = make(chan string, requestChanLength)
	hd.ResultChan = make(chan map[string]string, resultChanLength)
}

// 关闭通道
func (hd *HttpHandler) CloseChan() {
	close(hd.RequestChan)
	close(hd.ResultChan)
}

// 处理
func (hd *HttpHandler) SaveImages(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var urls []string
		err := parseBodySlice(r, &urls)
		if err != nil {
			log.Println(err)
			io.WriteString(w, err.Error())
			return
		}
		
		length := len(urls)
		hd.InitChan(length, length)
		defer hd.CloseChan()
		start := time.Now().Unix()
		
		for _, v := range urls {
			go func(v string) {
				_, saveUrl, err := getImg(v)
				if err != nil {
					log.Printf("save %s error, error:%v", saveUrl, err)
				} else {
					log.Printf("save %s success.", saveUrl)
				}
				hd.ResultChan <- map[string]string{v : saveUrl}
			}(v)
		}
		log.Printf("run goroutine num: %d", runtime.NumGoroutine())

		var resData []map[string]string
		for _ = range urls {
			select {
			case imgUrl := <-hd.ResultChan:
				resData = append(resData, imgUrl)
			}

		}
		end := time.Now().Unix()
		sub := end - start
		log.Printf("run goroutine num: %d", runtime.NumGoroutine())
		log.Printf("costTime: %d", sub)
		jsonData, _ := json.Marshal(resData)
		io.WriteString(w, string(jsonData))

	}
}
