package main

import (
	"errors"
	"fmt"
	"reflect"
)

type GenDownloader func() imgDownloader

type downloaderPool interface {
	Take() (imgDownloader, error)  // 取出
	Return(dl imgDownloader) error // 归还
	Total() uint32                 // 容量
	Used() uint32                  // 已用
}

func NewDownloaderPool(
	total uint32,
	gen GenDownloader) (downloaderPool, error) {
	etype := reflect.TypeOf(gen())
	// 使imgDownloader 转化为Entity接口类型
	genEntity := func() Entity {
		return gen()
	}
	pool, err := NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myDownloaderPool{
		pool:  pool,
		etype: etype,
	}
	return dlpool, nil
}

type myDownloaderPool struct {
	pool  Pool         // 实体池子
	etype reflect.Type // 池内实体的类型
}

func (dlpool *myDownloaderPool) Take() (imgDownloader, error) {
	entity, err := dlpool.pool.Take()
	if err != nil {
		return nil, err
	}
	dl, ok := entity.(imgDownloader)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is not %s!\n", dlpool.etype)
		panic(errors.New(errMsg))
	}
	return dl, nil
}

func (dlpool *myDownloaderPool) Return(dl imgDownloader) error {
	return dlpool.pool.Return(dl)
}

func (dlpool *myDownloaderPool) Total() uint32 {
	return dlpool.pool.Total()
}

func (dlpool *myDownloaderPool) Used() uint32 {
	return dlpool.pool.Used()
}
