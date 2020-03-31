/***********************************************************************
* @ 高性能队列 Disruptor
* @ 多生产者
	· 每个线程获取不同的一段数组空间进行操作
	· 引入了一个与Ring Buffer大小相同的buffer：available Buffer
	· 当某个位置写入成功的时候，便把availble Buffer相应的位置置位，标记为写入成功
	· 读取的时候，会遍历available Buffer，来判断元素是否已经就绪
* @ 读数据
	· 申请读取到序号n
	· 若writer cursor >= n，这时仍然无法确定连续可读的最大下标
	· 从reader cursor开始读取available Buffer，一直查到第一个不可用的元素，然后返回最大连续可读元素的位置；
* @ 写数据
	· 申请写入m个元素
	· 若是有m个元素可以写入，则返回最大的序列号，每个生产者会被分配一段独享的空间
	· 生产者写入元素，写入元素的同时设置available Buffer里面相应的位置，以标记哪些位置是已经写入成功的
***********************************************************************/
package safe

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

// 2019.3.28 https://github.com/yireyun/go-queue
// https://zhuanlan.zhihu.com/p/24432607

// 若goroutine会执行很长时间，且不是通过io阻塞或channel来同步，就需要主动调用Gosched()让出CPU
type Queue struct { //lock free queue
	kCap    uint32
	kCapMod uint32
	putPos  uint32
	getPos  uint32
	cache   []esCache
}
type esCache struct {
	putNo uint32
	getNo uint32
	value interface{}
}

func (q *Queue) Init(capaciity uint32) {
	q.kCap = minQuantity(capaciity)
	q.kCapMod = q.kCap - 1
	q.putPos = 0
	q.getPos = 0
	q.cache = make([]esCache, q.kCap)
	for i := range q.cache {
		cache := &q.cache[i]
		cache.getNo = uint32(i)
		cache.putNo = uint32(i)
	}
	cache := &q.cache[0]
	cache.getNo = q.kCap
	cache.putNo = q.kCap
}

func (q *Queue) Size() (putPos, getPos, size uint32) {
	getPos = atomic.LoadUint32(&q.getPos)
	putPos = atomic.LoadUint32(&q.putPos)
	if putPos >= getPos {
		size = putPos - getPos
	} else {
		size = q.kCapMod + (putPos - getPos)
	}
	return
}
func (q *Queue) Cap() uint32 { return q.kCap }

func (q *Queue) Put(val interface{}) (ok bool, size uint32) {
	var putPos uint32
	putPos, _, size = q.Size()
	capMod := q.kCapMod

	if size >= capMod-1 {
		runtime.Gosched()
		return false, size
	}
	putPosNew := putPos + 1
	if !atomic.CompareAndSwapUint32(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return false, size
	}
	cache := &q.cache[putPosNew&capMod]
	for {
		getNo := atomic.LoadUint32(&cache.getNo)
		putNo := atomic.LoadUint32(&cache.putNo)
		if putPosNew == putNo && getNo == putNo {
			cache.value = val
			atomic.AddUint32(&cache.putNo, q.kCap)
			return true, size + 1
		} else {
			runtime.Gosched()
		}
	}
}
func (q *Queue) Get() (val interface{}, ok bool, size uint32) {
	var getPos uint32
	_, getPos, size = q.Size()
	capMod := q.kCapMod

	if size < 1 {
		runtime.Gosched()
		return nil, false, size
	}
	getPosNew := getPos + 1
	if !atomic.CompareAndSwapUint32(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return nil, false, size
	}
	cache := &q.cache[getPosNew&capMod]
	for {
		getNo := atomic.LoadUint32(&cache.getNo)
		putNo := atomic.LoadUint32(&cache.putNo)
		if getPosNew == getNo && getNo == putNo-q.kCap {
			val = cache.value
			cache.value = nil
			atomic.AddUint32(&cache.getNo, q.kCap)
			return val, true, size - 1
		} else {
			runtime.Gosched()
		}
	}
}

// 批处理，建议大小是2N
func (q *Queue) Puts(ref []interface{}) (putCnt, size uint32) {
	var putPos uint32
	putPos, _, size = q.Size()
	capMod := q.kCapMod

	if size >= capMod-1 {
		runtime.Gosched()
		return 0, size
	}
	if capPuts, refCnt := q.kCap-size, uint32(len(ref)); capPuts >= refCnt {
		putCnt = refCnt
	} else {
		putCnt = capPuts
	}
	putPosNew := putPos + putCnt
	if !atomic.CompareAndSwapUint32(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return 0, size
	}
	for posNew, v := putPos+1, uint32(0); v < putCnt; posNew, v = posNew+1, v+1 {
		cache := &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint32(&cache.getNo)
			putNo := atomic.LoadUint32(&cache.putNo)
			if posNew == putNo && getNo == putNo {
				cache.value = ref[v]
				atomic.AddUint32(&cache.putNo, q.kCap)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return putCnt, size + putCnt
}
func (q *Queue) Gets(ref []interface{}) (getCnt, size uint32) {
	var getPos uint32
	_, getPos, size = q.Size()
	capMod := q.kCapMod

	if size < 1 {
		runtime.Gosched()
		return 0, size
	}
	if refCnt := uint32(len(ref)); size >= refCnt {
		getCnt = refCnt
	} else {
		getCnt = size
	}
	getPosNew := getPos + getCnt
	if !atomic.CompareAndSwapUint32(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return 0, size
	}
	for posNew, v := getPos+1, uint32(0); v < getCnt; posNew, v = posNew+1, v+1 {
		cache := &q.cache[posNew&capMod]
		for {
			getNo := atomic.LoadUint32(&cache.getNo)
			putNo := atomic.LoadUint32(&cache.putNo)
			if posNew == getNo && getNo == putNo-q.kCap {
				ref[v] = cache.value
				cache.value = nil
				getNo = atomic.AddUint32(&cache.getNo, q.kCap)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return getCnt, size - getCnt
}

// round 到最近的2的倍数
func minQuantity(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}
func (q *Queue) Print() string {
	return fmt.Sprintf("Queue{cap: %v, capMod: %v, putPos: %v, getPos: %v}",
		q.kCap, q.kCapMod, atomic.LoadUint32(&q.putPos), atomic.LoadUint32(&q.getPos))
}
