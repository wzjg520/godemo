package httphandler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"math/rand"
	"strconv"
	"time"
	"net"
)

// 解析http请求中的body体，这里用来解析json数组
func parseBodySlice(r *http.Request, m *[]string) error {
	var buf []byte
	//var buf2 bytes.Buffer
	//for {
	//	_, err := r.Body.Read(buf)
	//	buf2.Write(buf)
	//	fmt.Println(err)
	//	if err != nil || err == io.EOF {
	//		break
	//	}
	//}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	err = json.Unmarshal(bytes.TrimRight(buf, "\x00"), m)
	if err != nil {
		return err
	}
	return nil
}

// 解析http请求body中数据，机械map类型
func parseBodyMap(r *http.Request, m *map[string]interface{}) error {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	err = json.Unmarshal(bytes.TrimRight(buf, "\x00"), m)
	if err != nil {
		return err
	}
	return nil
}
// 根据图片地址下载图片
func getImg(url string) (n int64, saveUrl string, err error) {
	path := strings.Split(url, "/")
	if len(path) > 1 {
		saveUrl = strconv.Itoa(rand.Int()) + "_" + path[len(path)-1]
	}
	
	out, err := os.Create(saveUrl)
	if err != nil {
		return 0, "", err
	}
	defer out.Close()
	
	httpClient := TimeoutHttpClient(10 * time.Minute)
	
	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	n, err = io.Copy(out, bytes.NewReader(pix))
	if err != nil {
		return 0, "", err
	}
	return
}

func TimeoutHttpClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				return nil ,err
			}
			return c, nil
		},
	}
	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return &client
}
