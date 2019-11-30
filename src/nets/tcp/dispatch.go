// +build !tcp_multi

package tcp

import (
	"common"
	"gamelog"
)

func (self *RpcQueue) Insert(conn *TCPConn, msg *common.NetPack) {
	//self.queue.Put(objMsg{conn, msg})
	select {
	case self.queue <- objMsg{conn, msg}:
	default:
		gamelog.Error("RpcQueue Insert: channel full")
		conn.Close()
	}
}
