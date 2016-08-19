package itempipeline

import (
	"crawer/base"
	"errors"
	"fmt"
	"sync/atomic"
)

var summaryTemplate = "failFast:%v, proessorNumber: %d," +
	" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

// 条目处理器的接口类型
type ItemPipeline interface {
	// 发送条目
	Send(item base.Item) []error
	// FailFast 方法会返回一个bool值，该值表示当前的条目处理管道是否是快速失败的
	// 快速失败指：只要对某个条目的处理流程在某一个步骤上出错
	// 那么条目处理管道就会忽略掉后续的所有处理步骤并报告错误
	FailFast() bool
	// 设置是否快速失败
	SetFailFast(failFast bool)
	// 获得已发送、已接受和已处理的条目的计数值
	Count() []uint64
	// 获得正在被处理的条目的数量
	ProcessingNumber() uint64
	// 获得摘要信息
	Summary() string
}

// 创建条目处理管道
func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New(fmt.Sprintln("Invalid item processor list!")))
	}
	innerItemProcessors := make([]ProcessItem, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item processor[%d]!\n", i)))
		}
		innerItemProcessors = append(innerItemProcessors, ip)
	}
	return &myItemPipeline{itemProcessors: innerItemProcessors}
}

// 条目处理管道的实现类型
type myItemPipeline struct {
	itemProcessors   []ProcessItem // 条目处理器列表
	failFast         bool          // 是否快速失败
	sent             uint64        // 已被发送的条目的数量
	accepted         uint64        // 已被接收的条目的数量
	processed        uint64        // 已被处理的条目的数量
	processingNumber uint64        // 正被处理的条目的数量
}

// 发送并处理条目
func (ip *myItemPipeline) Send(item base.Item) []error {
	atomic.AddUint64(&ip.processingNumber, 1)
	defer atomic.AddUint64(&ip.processingNumber, ^uint64(0))
	atomic.AddUint64(&ip.sent, 1)
	errs := make([]error, 0)
	if item == nil {
		errs = append(errs, errors.New("The item is invalid!"))
		return errs
	}
	atomic.AddUint64(&ip.accepted, 1)
	var currentItem base.Item = item
	for _, itemProcessor := range ip.itemProcessors {
		processedItem, err := itemProcessor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if ip.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&ip.processed, 1)
	return errs

}

// 获得值
func (ip *myItemPipeline) FailFast() bool {
	return ip.failFast
}

func (ip *myItemPipeline) SetFailFast(failFast bool) {
	ip.failFast = failFast
}

func (ip *myItemPipeline) Count() []uint64 {
	counts := make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&ip.sent)
	counts[1] = atomic.LoadUint64(&ip.accepted)
	counts[2] = atomic.LoadUint64(&ip.processed)
	return counts
}

func (ip *myItemPipeline) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&ip.processingNumber)
}

func (ip *myItemPipeline) Summary() string {
	counts := ip.Count()
	summary := fmt.Sprintf(summaryTemplate,
		ip.failFast, len(ip.itemProcessors),
		counts[0], counts[1], counts[2], ip.ProcessingNumber())
	return summary
}
