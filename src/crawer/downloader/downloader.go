package downloader

import (
	"crawer/base"
	mdw "crawer/middleware"
	"logging"
	"net/http"
)

var logger logging.Logger = base.NewLogger()

var downloaderIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

func genDownloaderId() uint32 {
	return downloaderIdGenerator.GetUint32Id()
}

type PageDownloader interface {
	Id() uint32                                        // 获得Id
	Download(req base.Request) (*base.Response, error) // 根据请求下载网页并返回响应
}

type myPageDownloader struct {
	id         uint32      // id
	httpClient http.Client // http 客户端
}

func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &myPageDownloader{
		id:         id,
		httpClient: *client,
	}
}

func (dl *myPageDownloader) Id() uint32 {
	return dl.id
}

func (dl *myPageDownloader) Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()
	logger.Infof("Do the request (url=%s)... \n", httpReq.URL)
	httpResp, err := dl.httpClient.Do(httpReq)

	logger.Infoln(err)
	logger.Infoln(err)

	if err != nil {
		return nil, err
	}
	return base.NewResponse(httpResp, req.Depth()), nil
}
