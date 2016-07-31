package analyzer

import (
	mdw "crawer/middleware"
	"errors"
	"fmt"
	"reflect"
)

// 生成分析器的函数类型
type GenAnallyzer func() Analyzer

//分析器池的接口类型
type AnalyzerPool interface {
	Take() (Analyzer, error)        // 从池中取出一个分析器
	Return(analyzer Analyzer) error //把一个分析器归还给池
	Total() uint32                  //获得池的总容量
	Used() uint32                   // 获得正在被使用的分析器数量
}

type myAnalyzerPool struct {
	pool  mdw.Pool
	etype reflect.Type
}

func NewAnalyzerPool(total uint32, gen GenAnallyzer) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity {
		return gen()
	}
	pool, err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myAnalyzerPool{pool: pool, etype: etype}
	return dlpool, nil
}

func (mypool *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := mypool.pool.Take()
	if err != nil {
		return nil, err
	}
	analyzer, ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is not %s!\n", mypool.etype)
		panic(errors.New(errMsg))
	}
	return analyzer, nil

}

func (mypool *myAnalyzerPool) Return(analyzer Analyzer) error {
	return mypool.pool.Return(analyzer)
}

func (mypool *myAnalyzerPool) Total() uint32 {
	return mypool.pool.Total()
}

func (mypool *myAnalyzerPool) Used() uint32 {
	return mypool.pool.Used()
}
