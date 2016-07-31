package scheduler

import (
	"crawer/analyzer"
	anlz "crawer/analyzer"
	"crawer/base"
	dl "crawer/downloader"
	"crawer/itempipeline"
	ipl "crawer/itempipeline"
	mdw "crawer/middleware"
	"errors"
	"fmt"
	"logging"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
)

// 日志记录器
var logger logging.Logger = base.NewLogger()

// 被用来生成http客户端的函数类型
type GenHttpClient func() *http.Client

// 调度器接口类型
type Scheduler interface {
	Start(channelArgs base.ChannelArgs,
		poolBaseArgs base.PoolBaseArgs,
		crawDepth uint32,
		httpClientGenerator GenHttpClient,
		respParsers []analyzer.ParseResponse,
		itemProcessors []itempipeline.ProcessItem,
		firstHttpReq *http.Request) (err error)

	// 停止调度器运行，所有模块都会被终止
	Stop() bool
	// 判断调度器是否正在运行
	Running() bool
	// 获得错误通道
	ErrorChan() <-chan error

	// 判断所有处理模块是否都处于空闲状态
	Idle() bool
	// 获得摘要信息
	Summary(prefix string) SchedSummary
}

// 创建调度器
func NewScheduler() Scheduler {
	return &myScheduler{}
}

type myScheduler struct {
	channelArgs   base.ChannelArgs      // 通道参数的容器
	poolBaseArgs  base.PoolBaseArgs     // 池基本参数的容器
	crawlDepth    uint32                // 爬取的最大深度
	primaryDomain string                // 主域名
	chanman       mdw.ChannelManager    // 通道管理器
	stopSign      mdw.StopSign          // 停止信号
	dlpool        dl.PageDownloaderPool // 网页下载器池
	analyzerPool  anlz.AnalyzerPool     // 分析器池
	itemPipeline  ipl.ItemPipeline      // 条目处理管道
	reqCache      requestCache          // 已请求的url的字典
	urlMap        map[string]bool       // 已请求的url的字典
	running       uint32                // 运行标记0 未运行 1已运行 2已停止

}

func (sched *myScheduler) Start(
	channelArgs base.ChannelArgs,
	poolBaseArgs base.PoolBaseArgs,
	crawDepth uint32,
	httpClientGenerator GenHttpClient,
	respParsers []anlz.ParseResponse,
	itemProcessors []ipl.ProcessItem,
	firstHttpReq *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Scheduler Error: %s\n", p)
			logger.Fatal(errMsg)
			err = errors.New(errMsg)
		}

	}()

	if atomic.LoadUint32(&sched.running) == 1 {
		return errors.New("The scheduler has been started!\n")
	}

	atomic.StoreUint32(&sched.running, 1)

	if err := channelArgs.Check(); err != nil {
		return err
	}
	sched.channelArgs = channelArgs

	if err := poolBaseArgs.Check(); err != nil {
		return err
	}
	sched.poolBaseArgs = poolBaseArgs
	sched.crawlDepth = crawDepth
	sched.chanman = generateChannelManager(sched.channelArgs)
	if httpClientGenerator == nil {
		return errors.New("The http client generator list is invalid!")
	}
	dlpool, err := generatePageDownloaderPool(sched.poolBaseArgs.PageDownloaderPoolSize(),
		httpClientGenerator)

	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page download pool:%s\n", err)
		return errors.New(errMsg)
	}

	sched.dlpool = dlpool
	analyzerPool, err := generateAnalyzerPool(sched.poolBaseArgs.AnalyzerPoolSize())

	if err != nil {
		errMsg := fmt.Sprintf("occur error when get analyzer pool:%s\n", err)
		return errors.New(errMsg)
	}

	sched.analyzerPool = analyzerPool
	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!")
	}

	for i, ip := range itemProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}
	sched.itemPipeline = generateItemPipeline(itemProcessors)
	if sched.stopSign == nil {
		sched.stopSign = mdw.NewStopSign()
	} else {
		sched.stopSign.Reset()
	}

	sched.reqCache = newRequestCache()
	sched.urlMap = make(map[string]bool)

	sched.startDownloading()
	sched.activateAnalyzers(respParsers)
	sched.openItemPipeline()
	sched.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The first Http request is invalid!")
	}
	pd, err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	sched.primaryDomain = pd

	firstReq := base.NewRequest(firstHttpReq, 0)
	sched.reqCache.put(firstReq)
	return nil
}

func (sched *myScheduler) Stop() bool {
	if atomic.LoadUint32(&sched.running) != 1 {
		return false
	}
	sched.stopSign.Sign()
	sched.chanman.Close()
	sched.reqCache.close()
	atomic.StoreUint32(&sched.running, 2)
	return true
}

func (sched *myScheduler) Running() bool {
	return atomic.LoadUint32(&sched.running) == 1
}

func (sched *myScheduler) ErrorChan() <-chan error {
	if sched.chanman.Status() != mdw.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrorChan()
}

func (sched *myScheduler) Idle() bool {
	usedDlPool := sched.dlpool.Used() == 0
	usedAyPool := sched.analyzerPool.Used() == 0
	usedItPipeline := sched.itemPipeline.ProcessingNumber() == 0
	if usedDlPool && usedAyPool && usedItPipeline {
		return true
	}
	return false
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(sched, prefix)
}

// 开始下载
func (sched *myScheduler) startDownloading() {

	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				break
			}
			go sched.download(req)
		}
	}()
}

// 下载
func (sched *myScheduler) download(req base.Request) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Download Error:%s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	downloader, err := sched.dlpool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Downloader pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.dlpool.Return(downloader)
		if err != nil {
			errMsg := fmt.Sprintf("Downloader pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()

	code := generateCode(DOWNLOADER_CODE, downloader.Id())
	resp, err := downloader.Download(req)
	if resp != nil {
		sched.sendResp(*resp, code)
	}
	if err != nil {
		sched.sendError(err, code)
	}
}

// 激活分析器
func (sched *myScheduler) activateAnalyzers(respParsers []anlz.ParseResponse) {
	go func() {
		for {
			resp, ok := <-sched.getRespChan()
			if !ok {
				break
			}
			go sched.analyze(respParsers, resp)
		}
	}()
}

// 分析
func (sched *myScheduler) analyze(respParses []anlz.ParseResponse, resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis  Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	analyzer, err := sched.analyzerPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.analyzerPool.Return(analyzer)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()

	code := generateCode(ANALYZER_CODE, analyzer.Id())
	dataList, errs := analyzer.Analyze(respParses, resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *base.Request:
				sched.saveReqToCache(*d, code)
			case *base.Item:
				sched.sendItem(*d, code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", d, d)
				sched.sendError(errors.New(errMsg), code)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sched.sendError(err, code)
		}
	}

}

// 打开条目处理管道
func (sched *myScheduler) openItemPipeline() {
	go func() {
		sched.itemPipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil {
						errMsg := fmt.Sprintf("Fatal item Processing Error:%s\n", p)
						logger.Fatal(errMsg)
					}
				}()
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						sched.sendError(err, code)
					}
				}
			}(item)
		}
	}()
}

// 把请求放到请求缓存
func (sched *myScheduler) saveReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		logger.Warnln("Ignore the request! this http is invalid!")
		return false
	}
	reqUrl := httpReq.URL

	if reqUrl == nil {
		logger.Warnln("Ignore the request! this url is invalid!")
		return false
	}

	if strings.ToLower(reqUrl.Scheme) != "http" {
		logger.Warnf("Ignore the request! this scheme '%s', bug should be 'http'!\n", reqUrl.Scheme)
		return false
	}

	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		logger.Warnf("Ignore the request! this url is repeated. (requestUrl=%s)\n", reqUrl)
	}

	if pd, _ := getPrimaryDomain(httpReq.Host); pd != sched.primaryDomain {
		logger.Warnf("Ignore the request! this host '%s' not in primary domain '%s'. (requestUrl=%s)\n", httpReq.Host, sched.primaryDomain, reqUrl)

	}

	if req.Depth() > sched.crawlDepth {
		logger.Warnf("Ignore the request! this depth %d greater than %d. (requestUrl=%s)\n",
			req.Depth(), sched.crawlDepth, reqUrl)
		return false
	}

	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.reqCache.put(&req)
	sched.urlMap[reqUrl.String()] = true
	return true

}

// 发送响应
func (sched *myScheduler) sendResp(resp base.Response, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp
	return true
}

// 发送条目
func (sched *myScheduler) sendItem(item base.Item, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getItemChan() <- item
	return true
}

// 发送错误
func (sched *myScheduler) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType base.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = base.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = base.ITEM_PROCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errorType, err.Error())

	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	go func() {
		sched.getErrorChan() <- cError
	}()
	return true
}

// 调度
func (sched *myScheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.Signed() {
				sched.stopSign.Deal(SCHEDULER_CODE)
				return
			}
			remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
			var temp *base.Request
			for remainder > 0 {
				temp = sched.reqCache.get()
				if temp == nil {
					break
				}
				if sched.stopSign.Signed() {
					sched.stopSign.Deal(SCHEDULER_CODE)
					return
				}
				sched.getReqChan() <- *temp
				remainder--
			}
		}
		time.Sleep(interval)
	}()
}

// 获取通道管理器持有的请求通道
func (sched *myScheduler) getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

// 获取响应通道
func (sched *myScheduler) getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan

}

// 获取通道管理器持有的条目通道
func (sched *myScheduler) getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}

// 获取通道管理器持有的错误通道
func (sched *myScheduler) getErrorChan() chan error {
	errorChan, err := sched.chanman.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errorChan
}
