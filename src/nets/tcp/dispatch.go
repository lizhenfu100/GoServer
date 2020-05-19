// +build !net_multi

package tcp

import (
	"bufio"
	"common"
	"common/assert"
	"encoding/binary"
	"gamelog"
	"io"
	"nets/rpc"
)

type netbuf struct{ *bufio.Reader }

func (b *netbuf) Init(r io.Reader) { b.Reader = bufio.NewReader(r) }

func (self *TCPConn) readLoop() {
	var err error
	var msgLen int
	for msgHead, q, r := make([]byte, kHeadLen), &rpc.G_RpcQueue, self.reader.Reader; ; {
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

var _sendBuf = common.NewNetPackCap(64) //单线程CallRpc复用

//Notice：非线程安全的，仅供主逻辑线程调用
func (self *TCPConn) CallEx(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	req := _sendBuf
	assert.True(req.GetMsgId() == 0) //中途不能再CallRpc，同个sendBuf被覆盖
	rpc.MakeReq(req, msgId, sendFun, recvFun)
	self.WriteMsg(req)
	req.SetMsgId(0)
	req.ClearBody()
}
