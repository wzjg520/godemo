package main

import (
	"math"
	"sync"
)

// ID生成器接口类型
type IdGenerator interface {
	GetUint32Id() uint32
}

// ID循环生成器的实现类型
type cyclicIdGenerator struct {
	sn    uint32     // 当前ID
	ended bool       // 前一个Id是否已经为其类型所能表示的最大值
	mutex sync.Mutex // 互斥锁
}

func NewIdGenerator() IdGenerator {
	return &cyclicIdGenerator{}
}

func (gen *cyclicIdGenerator) GetUint32Id() uint32 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	// 前一个id已经为最大值时，重新生成一个
	if gen.ended {
		defer func() { gen.ended = false }()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	} else {
		gen.ended = true
	}
	return id
}

// ID生成器的接口实现类型2
type IdGenerator2 interface {
	GetUint64Id() uint64 // 获得一个uint64类型的ID
}

type cyclicIdGenerator2 struct {
	base       cyclicIdGenerator // 基本的ID生成器
	cycleCount uint64            // 基于uint32类型的取值范围的周期计数
}

func NewIdGenerator2() IdGenerator2 {
	return &cyclicIdGenerator2{}
}

func (gen *cyclicIdGenerator2) GetUint64Id() uint64 {
	var id64 uint64
	if gen.cycleCount%2 == 1 {
		id64 += math.MaxInt32
	}
	id32 := gen.base.GetUint32Id()

	if id32 == math.MaxInt32 {
		gen.cycleCount++
	}
	id64 += uint64(id32)
	return id64

}
