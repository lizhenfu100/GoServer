/***********************************************************************
* @ 临时对象池，按尺寸划分
* @ brief
	1、我们可以把sync.Pool类型值看作是存放可被重复使用的值的容器，自动伸缩、高效、并发安全

	2、它会专门为每一个与操作它的goroutine相关联的Pool都生成一个本地池。

	3、在临时对象池的Get方法被调用的时候，它一般会先尝试从与本地Pool对应的那个本地池中获取一个对象值。
		如果获取失败，它就会试图从其他Pool的本地池中偷一个对象值并直接返回给调用方。
		如果依然未果，那它只能把希望寄托于当前的临时对象池的New字段代表的那个对象值生成函数了。
		注意，这个对象值生成函数产生的对象值永远不会被放置到池中。它会被直接返回给调用方。

	4、临时对象池的Put方法会把它的参数值存放到与当前P对应的那个本地池中。
		每个P的本地池中的绝大多数对象值都是被同一个临时对象池中的所有本地池所共享的。也就是说，它们随时可能会被偷走

	5、对gc友好，gc执行时临时对象池中的某个对象值仅被该池引用，那么它可能会在gc时被回收

* @ Notice
	1、pool包在init的时候注册了一个poolCleanup函数，它会清除所有的pool里面的所有缓存的对象
		该函数注册进去之后会在每次gc之前都会调用
		因此sync.Pool缓存的期限只是两次gc之间这段时间

	2、不能控制Pool中的元素数量，放进Pool中的对象每次GC发生时可能都会被清理掉

* @ author 达达
* @ date 2016-7-23
************************************************************************/
package pool

import "sync"

type SyncPool []syncPool
type syncPool struct {
	sync.Pool
	size int
}

// create a sync.Pool base slab allocation memory pool.
// minSize is the smallest chunk size.
// maxSize is the lagest chunk size.
// factor is used to control growth of chunk size.
// pool.Init(128, 1024, 2)
func (self *SyncPool) Init(minSize, maxSize, factor int) {
	for size := minSize; size <= maxSize; size *= factor {
		v := syncPool{size: size}
		v.New = func() interface{} { return make([]byte, v.size) }
		*self = append(*self, v)
	}
}
func (self *SyncPool) Alloc(nLen int) []byte {
	for i := 0; i < len(*self); i++ {
		if p := (*self)[i]; p.size >= nLen {
			return p.Get().([]byte)[:nLen]
		}
	}
	return make([]byte, nLen)
}
func (self *SyncPool) Free(buf []byte) {
	for i := 0; i < len(*self); i++ {
		if p := (*self)[i]; p.size >= cap(buf) {
			p.Put(buf)
			return
		}
	}
}
