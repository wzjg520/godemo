package analyzer

import (
	"crawer/base"
	"net/http"
)

// 解析htt响应的函数类型
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)
