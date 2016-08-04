package httphandler

import (
	"net/http"
	//"encoding/json"
	"io"
	"log"
)

type HttpHandler struct {
	RequestChan chan string
	ResultChan  chan string
}

// 生成资源对象
func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

// 初始化通道
func (hd *HttpHandler) Init(requestChanLength int, resultChanLength int) {
	hd.RequestChan = make(chan string, requestChanLength)
	hd.ResultChan = make(chan string, resultChanLength)
}

// 处理
func (hd *HttpHandler) SaveImages(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var jsonObj []string
		err := parseBodySlice(r, &jsonObj)
		if err != nil {
			log.Println(err)
			io.WriteString(w, err.Error())
			return
		}
		//json, _ := json.Marshal(jsonObj)

		for _, v := range jsonObj {
			go func() {
				log.Println(v + "\n")
				getImg(v)
				hd.ResultChan <- (v + "\n")
			}()
		}

		str := ""
		for _ = range jsonObj {
			select {
			case imgUrl := <-hd.ResultChan:
				str = imgUrl
			}

		}

		io.WriteString(w, str)

	}
}
