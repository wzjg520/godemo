package httphandler

import (
	"net/http"
	//"encoding/json"
	"encoding/json"
	"io"
	"log"
	"runtime"
	"time"
)

type HttpHandler struct {
	cacheDir    string
	scriptPath  string
	RequestChan chan string
	ResultChan  chan map[string]string
}

// 生成资源对象
func NewHttpHandler(cacheDir string, scriptPath string) *HttpHandler {
	return &HttpHandler{
		cacheDir:   cacheDir,
		scriptPath: scriptPath,
	}
}

// 初始化通道
func (hd *HttpHandler) InitChan(requestChanLength int, resultChanLength int) {
	log.Printf("init resultChan length %d", resultChanLength)
	log.Printf("init a requestChan length %d", requestChanLength)

	if (resultChanLength > 50) {
		hd.ResultChan = make(chan map[string]string, 50)
	} else {
		hd.ResultChan = make(chan map[string]string, resultChanLength)
	}

	if (requestChanLength > 50) {
		hd.RequestChan = make(chan string, 50)
	} else {
		hd.RequestChan = make(chan string, requestChanLength)
	}

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
				defer func() {
					if p := recover(); p != nil {
						log.Println(p)
					}
				}()
				_, saveUrl, err := getImg(v, hd.cacheDir, hd.scriptPath)

				if err != nil {
					log.Printf("save %s error, error:%v", saveUrl, err)
				} else {
					log.Printf("save %s success.", saveUrl)
				}
				hd.ResultChan <- map[string]string{v: saveUrl}
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
