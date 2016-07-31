package base

import (
	"errors"
	"fmt"
)

// 参数容器接口
type Args interface {
	Check() error
	String() string
}

// 通道参数容器描述模板
var channelArgsTemplate string = "{ reqCanLen: %d, respChanLen: %d" +
	" itemChanLen: %d, errorChanLen: %d }"

// 通道参数的容器
type ChannelArgs struct {
	reqChanLen   uint
	respChanLen  uint
	itemChanLen  uint
	errorChanLen uint
	description  string
}

func (args *ChannelArgs) Check() error {
	if args.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if args.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")

	}
	if args.itemChanLen == 0 {
		return errors.New("The item channel max length (capacity) can not be 0!\n")
	}
	if args.errorChanLen == 0 {
		return errors.New("The error Channel max length (capacity) can not be 0!\n")
	}

	return nil
}

func (args *ChannelArgs) String() string {
	if args.description == "" {
		args.description = fmt.Sprintf(channelArgsTemplate,
			args.reqChanLen,
			args.respChanLen,
			args.itemChanLen,
			args.errorChanLen)
	}

	return args.description
}

func (args *ChannelArgs) ReqChanLen() uint {
	return args.reqChanLen
}

func (args *ChannelArgs) RespChanLen() uint {
	return args.reqChanLen
}

func (args *ChannelArgs) ItemChanLen() uint {
	return args.itemChanLen
}

func (args *ChannelArgs) ErrorChanLen() uint {
	return args.errorChanLen
}

func NewChannelArgs(
	reqChanLen uint,
	respChanLen uint,
	itemChanLen uint,
	errorChanLen uint) ChannelArgs {
	return ChannelArgs{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		itemChanLen:  itemChanLen,
		errorChanLen: errorChanLen,
	}
}

// 池子基本参数容器描述模板
var poolBaseArgsTemplate string = "{ pageDownloaderPoolSize: %d," +
	" analyzerPoolSize: %d }"

// 池子基本参数的容器
type PoolBaseArgs struct {
	pageDownloaderPoolSize uint32
	analyzerPoolSize       uint32
	description            string
}

func NewPoolBaseArgs(pageDownloaderPoolSize uint32, analyzerPoolSize uint32) PoolBaseArgs {
	return PoolBaseArgs{
		pageDownloaderPoolSize: pageDownloaderPoolSize,
		analyzerPoolSize:       analyzerPoolSize,
	}
}

func (args *PoolBaseArgs) Check() error {
	if args.pageDownloaderPoolSize == 0 {
		return errors.New("The page downloader pool size can not be 0!\n")
	}
	if args.analyzerPoolSize == 0 {
		return errors.New("The page analyzer pool size can not be 0!\n")
	}
	return nil

}

func (args *PoolBaseArgs) String() string {
	if args.description == "" {
		args.description = fmt.Sprintf(poolBaseArgsTemplate,
			args.pageDownloaderPoolSize,
			args.analyzerPoolSize)
	}
	return args.description
}

func (args *PoolBaseArgs) PageDownloaderPoolSize() uint32 {
	return args.pageDownloaderPoolSize
}

func (args *PoolBaseArgs) AnalyzerPoolSize() uint32 {
	return args.analyzerPoolSize
}
