package httphandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 解析http请求中的body体，这里用来解析json数组
func parseBodySlice(r *http.Request, m *[]string) error {
	var buf []byte

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

// 解析http请求body中数据，解析map类型
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
func getImg(url string, cacheDir string, scriptPath string) (n int64, saveUrl string, err error) {
	ext := strings.Trim(filepath.Ext(url), ".")

	imgExtMap := map[string]bool{
		"jpg":  true,
		"png":  true,
		"jpeg": true,
		"gif":  true,
	}

	if _, ok := imgExtMap[ext]; !ok {
		return 0, "", errors.New(fmt.Sprintf("%s not a image!", url))
	}

	path := strings.Split(url, "/")

	if len(path) > 1 {
		saveUrl = filepath.Join(cacheDir, strconv.Itoa(rand.Int())+"_"+path[len(path)-1])
	}

	out, err := os.Create(saveUrl)
	if err != nil {
		return 0, "", err
	}

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
	out.Close()
	saveUrl, err = execPhpScript(saveUrl, scriptPath)
	return
}

// http 请求设置超时
func TimeoutHttpClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				return nil, err
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

// 执行php脚本
func execPhpScript(url string, scriptPath string) (saveUrl string, err error) {
	if strings.TrimSpace(url) == "" {
		return "", nil
	}
	data, err := exec.Command("php", "-f", scriptPath, url).Output()
	if err != nil {
		log.Println(err)
	}

	return string(data), err
}
