package qps

import (
	"gamelog"
	"sync/atomic"
	"time"
)

var (
	g_recvNum uint32
)

func WatchLoop() {
	for range time.Tick(time.Second) {
		gamelog.Info("QPS: %d", atomic.SwapUint32(&g_recvNum, 0))
	}
}

func AddQps() { atomic.AddUint32(&g_recvNum, 1) }
