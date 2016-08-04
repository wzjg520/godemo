package httphandler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"math/rand"
	"strconv"
)

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

func getImg(url string) (n int64, err error) {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = strconv.Itoa(rand.Int()) + "_" + path[len(path)-1]
	}
	log.Println(name)
	out, err := os.Create(name)
	defer out.Close()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	n, err = io.Copy(out, bytes.NewReader(pix))
	return
}
