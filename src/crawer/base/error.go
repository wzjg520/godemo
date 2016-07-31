package base

import (
	"bytes"
	"fmt"
)

// 错误类型
type ErrorType string

// 错误类型常量
const (
	DOWNLOADER_ERROR     ErrorType = "Downloader Error"
	ANALYZER_ERROR       ErrorType = "Analyzer Error"
	ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

// 爬虫错误的接口
type CrawlerError interface {
	Type() ErrorType // 获得错误类型
	Error() string   // 获得错误提示信息
}

// 爬虫错误的实现
type myCrawlerError struct {
	errorType    ErrorType // 错误类型
	errMsg       string    // 错误提示信息
	fullErrorMsg string    // 完整的错误提示信息
}

func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{errorType: errType, errMsg: errMsg}
}

func (ce *myCrawlerError) Type() ErrorType {
	return ce.errorType
}

func (ce *myCrawlerError) Error() string {
	if ce.fullErrorMsg == "" {
		ce.genFullErrMsg()
	}

	return ce.fullErrorMsg
}

func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error:")
	if ce.errorType != "" {
		buffer.WriteString(string(ce.errorType))
		buffer.WriteString(": ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrorMsg = fmt.Sprintf("%s\n", buffer.String())
	return
}
