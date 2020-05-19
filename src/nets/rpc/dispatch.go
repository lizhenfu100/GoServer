// +build !net_multi

package rpc

import (
	"common"
	"gamelog"
)

const Msg_Queue_Cap = 10240

var _queue = make(chan objMsg, Msg_Queue_Cap) //safe.Pipe

func (self *RpcQueue) Insert(conn common.Conn, msg *common.NetPack) {
	select {
	case _queue <- objMsg{conn, msg}:
	default:
		gamelog.Error("RpcQueue Insert: channel full")
		conn.Close()
	}
}
func (self *RpcQueue) Update() { //主循环，每帧调一次
	for ack := common.NewNetPackCap(64); ; {
		select {
		case v := <-_queue:
			self.Handle(v.conn, v.msg, ack)
			v.msg.Free()
		default:
			return
		}
	}
}
