package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// 实体的接口类型
type Entity interface {
	Id() uint32
}

// goroutine池接口类型
type Pool interface {
	Take() (Entity, error)      // 取出实体
	Return(entity Entity) error // 归还实体
	Total() uint32              // 实体池的总容量
	Used() uint32               // 实体池中已经使用的实体
}

// 创建实体池子
func NewPool(total uint32, entityType reflect.Type, genEntity func() Entity) (Pool, error) {
	if total == 0 {
		errMsg := fmt.Sprintf("The pool can not be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	size := int(total)
	container := make(chan Entity, size)
	idContainer := make(map[uint32]bool)
	for i := 0; i < size; i++ {
		newEntity := genEntity()
		if entityType != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("The type of result of function genEntity() is not %s!\n", entityType)
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}

	pool := &MyPool{
		total:       total,
		etype:       entityType,
		genEntity:   genEntity,
		container:   container,
		idContainer: idContainer,
	}
	return pool, nil
}

// 池子实现
type MyPool struct {
	total       uint32          // 池的总容量
	etype       reflect.Type    // 池中实体的类型
	genEntity   func() Entity   // 实体生成函数
	container   chan Entity     // 实体容器
	idContainer map[uint32]bool // 实体id容器
	mutex       sync.Mutex      // 实体id操作互斥锁
}

// 取出一个实体
func (pool *MyPool) Take() (Entity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *MyPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}
	if pool.etype != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is not %s!\n", pool.etype)
		return errors.New(errMsg)
	}
	entityId := entity.Id()
	caseResutl := pool.compareAndSetForIdContainer(entityId, false, true)
	switch caseResutl {
	case 1:
		pool.container <- entity
		return nil
	case 0:
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entityId)
		return errors.New(errMsg)
	default:
		errMsg := fmt.Sprintf("The entity (id=%d) is illegal!\n", entityId)
		return errors.New(errMsg)
	}
}

// -1 : 键值不存在
// 	0 : 操作失败
//  1 : 操作成功
func (pool *MyPool) compareAndSetForIdContainer(entityId uint32, oldValue bool, newValue bool) int8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	v, ok := pool.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	pool.idContainer[entityId] = newValue
	return 1
}

func (pool *MyPool) Total() uint32 {
	return pool.total
}

func (pool *MyPool) Used() uint32 {
	return pool.total - uint32(len(pool.container))
}
