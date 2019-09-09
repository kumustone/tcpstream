package tcpstream

import "sync/atomic"

type Int struct {
	id uint64
}

//在一个低频系统中，忽略因为溢出导致的ID重复, 在业务层面可以不同的模块使用自己单独的ID生成器；
func (i *Int) GenID() uint64 {
	return atomic.AddUint64(&(i.id), 1)
}
