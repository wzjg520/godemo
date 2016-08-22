package main

import "log"

var downloaderIdGenerator = NewIdGenerator()

func genDownloaderId() uint32 {
	return downloaderIdGenerator.GetUint32Id()
}

type imgDownloader interface {
	Id() uint32
	SaveImages(url string) map[string]string
}

func NewDownloader() imgDownloader {
	return &downloader{
		id:         genDownloaderId(),
		cacheDir:   *cacheDir,
		scriptPath: *scriptPath,
	}
}

type downloader struct {
	id         uint32
	cacheDir   string
	scriptPath string
}

// 处理
func (dp *downloader) SaveImages(url string) map[string]string {

	_, saveURL, err := getImg(url, dp.cacheDir, dp.scriptPath)

	if err != nil {
		log.Printf("save %s error, error:%v", saveURL, err)
	} else {
		log.Printf("save %s success.", saveURL)
	}
	return map[string]string{url: saveURL}
}

func (dp *downloader) Id() uint32 {
	return dp.id
}
