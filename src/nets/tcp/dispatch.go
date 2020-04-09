// +build !tcp_multi

package tcp

import (
	"bufio"
	"common"
	"encoding/binary"
	"gamelog"
	"io"
)

func (self *RpcQueue) Insert(conn *TCPConn, msg *common.NetPack) {
	select {
	case self.queue <- objMsg{conn, msg}:
	default:
		gamelog.Error("RpcQueue Insert: channel full")
		conn.Close()
	}
}
func (self *RpcQueue) Update() { //主循环，每帧调一次
	for {
		select {
		case v := <-self.queue:
			self._Handle(v.conn, v.msg, self.backBuf)
			v.msg.Free()
		default:
			return
		}
	}
}

type netbuf struct{ *bufio.Reader }

func (b *netbuf) Init(r io.Reader) { b.Reader = bufio.NewReader(r) }

func (self *TCPConn) readLoop() {
	var err error
	var msgLen int
	for msgHead, q, r := make([]byte, kHeadLen), &G_RpcQueue, self.reader.Reader; ; {
		if self.IsClose() {
			break
		}
		if _, err = io.ReadAtLeast(r, msgHead, kHeadLen); err != nil {
			gamelog.Debug(err.Error())
			break
		}
		msgLen = int(binary.LittleEndian.Uint16(msgHead))
		if msgLen < common.PACK_HEADER_SIZE || msgLen > Msg_Size_Max {
			gamelog.Error("invalid msgLen: %d", msgLen)
			break
		}
		req := common.NewByteBufferLen(msgLen)
		req.ReadPos = common.PACK_HEADER_SIZE
		if _, err = io.ReadAtLeast(r, req.Data(), msgLen); err != nil {
			break
		}
		//FIXME: 消息加密、验证有效性，放逻辑线程，io线程只管io
		q.Insert(self, req) //转主线程处理消息，无竞态
	}
	self.Close()
	self.writer.Stop() //stop writeLoop
}
