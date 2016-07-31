package base

import ()
import "net/http"

// 数据接口
type Data interface {
	Valid() bool // 数据是否有效
}

// 请求
type Request struct {
	httpReq *http.Request
	depth   uint32
}

// 创建新的请求
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

// 获取Http请求
func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}

// 获取深度值
func (req *Request) Depth() uint32 {
	return req.depth
}

// 请求是否有效
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

// 响应
type Response struct {
	httpResp *http.Response
	depth    uint32
}

// 创建新的响应
func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

func (resp *Response) HttpResp() *http.Response {
	return resp.httpResp
}

// 获取深度值
func (resp *Response) Depth() uint32 {
	return resp.depth
}

// 数据是否有效
func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}

// 条目
type Item map[string]interface{}

// 数据是否有效
func (item Item) Valid() bool {
	return item != nil
}
