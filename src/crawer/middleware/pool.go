package middleware

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

// 实体池的接口类型
type Pool interface {
	Take() (Entity, error)      // 取出实体
	Return(entity Entity) error // 归还实体
	Total() uint32              // 实体池的容量
	Used() uint32               // 实体池中已使用的实体
}

// 创建实体池
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
	pool := &myPool{
		total:       total,
		etype:       entityType,
		genEntity:   genEntity,
		container:   container,
		idContainer: idContainer,
	}
	return pool, nil
}

type myPool struct {
	total       uint32          // 池的总容量
	etype       reflect.Type    // 池中实体的类型
	genEntity   func() Entity   // 实体生成函数
	container   chan Entity     // 实体容器
	idContainer map[uint32]bool // 实体id容器
	mutex       sync.Mutex      // 实体Id容器操作的互斥锁
}

func (pool *myPool) Take() (Entity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *myPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}
	if pool.etype != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is not %s!\n", pool.etype)
		return errors.New(errMsg)
	}
	entityId := entity.Id()
	casResult := pool.compareAndSetForIdContainer(entityId, false, true)
	switch casResult {
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

// 比较并设置实体id容器中与给定实体id对应的键值对元素值
// 结果
// 		-1 : 键值对不存在
//		 0 : 表示操作失败
//		 1 : 表示操作成功
func (pool *myPool) compareAndSetForIdContainer(entityId uint32, oldValue bool, newValue bool) int8 {
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

func (pool *myPool) Total() uint32 {
	return pool.total
}

func (pool *myPool) Used() uint32 {
	return pool.total - uint32(len(pool.container))
}
