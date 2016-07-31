package scheduler

import (
	"crawer/base"
	"fmt"
	"sync"
)

var statusMap = map[byte]string{
	0: "running",
	1: "closed",
}

var summaryTemplate = "status: %s, " + "length: %d, " + "capacity: %d"

// 请求缓存的接口类型
type requestCache interface {
	// 将请求放入请求缓存
	put(req *base.Request) bool
	get() *base.Request
	// 获得请求缓存的容量
	capacity() int

	length() int

	close()

	summary() string
}

func newRequestCache() requestCache {
	rc := &reqCacheBySlice{
		cache: make([]*base.Request, 0),
	}
	return rc
}

type reqCacheBySlice struct {
	cache  []*base.Request // 请求的存储介质
	mutex  sync.Mutex
	status byte // 缓存状态 0 正在运行 1 已关闭
}

func (rc *reqCacheBySlice) put(req *base.Request) bool {
	if req == nil {
		return false
	}

	if rc.status == 1 {
		return false
	}
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.cache = append(rc.cache, req)
	return true

}

func (rc *reqCacheBySlice) get() *base.Request {
	if rc.length() == 0 {
		return nil
	}
	if rc.status == 1 {
		return nil
	}
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	req := rc.cache[0]
	rc.cache = rc.cache[1:]
	return req
}

func (rc *reqCacheBySlice) length() int {
	return len(rc.cache)
}

func (rc *reqCacheBySlice) capacity() int {
	return cap(rc.cache)
}

func (rc *reqCacheBySlice) close() {
	if rc.status == 1 {
		return
	}
	rc.status = 1
}

func (rc *reqCacheBySlice) summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		statusMap[rc.status],
		rc.length(),
		rc.capacity())
	return summary
}
