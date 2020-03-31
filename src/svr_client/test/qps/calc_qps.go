package qps

import (
	"common/timer"
	"gamelog"
	"sync/atomic"
)

var g_recvNum uint32

func Watch() {
	timer.AddTimer(func() {
		gamelog.Info("QPS: %d", atomic.SwapUint32(&g_recvNum, 0))
	}, 1, 1, -1)
}
func AddQps() { atomic.AddUint32(&g_recvNum, 1) }
