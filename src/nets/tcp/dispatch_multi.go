// +build tcp_multi

package tcp

import (
	"common"
	"encoding/binary"
	"gamelog"
	"io"
)

func (self *RpcQueue) Insert(conn *TCPConn, req *common.NetPack) {
	if msgFunc := G_HandleFunc[req.GetMsgId()]; msgFunc != nil {
		msgFunc(req, nil, conn)
	}
}
func (self *TCPConn) readLoop() {
	var err error
	var msgLen int
	var req, ack common.NetPack
	ack.Reset(make([]byte, common.PACK_HEADER_SIZE, 32), common.PACK_HEADER_SIZE)
	for msgBuf, q := make([]byte, Msg_Size_Max), &G_RpcQueue; ; {
		if self.IsClose() {
			break
		}
		if _, err = io.ReadAtLeast(self.reader, msgBuf, kHeadLen); err != nil {
			break
		}
		msgLen = int(binary.LittleEndian.Uint16(msgBuf))
		if msgLen < common.PACK_HEADER_SIZE || msgLen > Msg_Size_Max {
			gamelog.Error("invalid msgLen: %d", msgLen)
			break
		}
		if _, err = io.ReadAtLeast(self.reader, msgBuf[kHeadLen:], msgLen); err != nil {
			break
		}
		//FIXME: 消息加密、验证有效性，pipeline模式优化
		req.Reset(msgBuf[kHeadLen:kHeadLen+msgLen], common.PACK_HEADER_SIZE)
		q._Handle(self, &req, &ack) //io线程直接处理，须考虑竞态
	}
	self.Close()
}
