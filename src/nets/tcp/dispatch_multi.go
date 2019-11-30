// +build tcp_multi

package tcp

import (
	"common"
)

func (self *RpcQueue) Insert(conn *TCPConn, msg *common.NetPack) {
	backBuf := common.NewNetPackCap(128)
	self._Handle(conn, msg, backBuf) //io线程直接处理消息
	backBuf.Free()
}
