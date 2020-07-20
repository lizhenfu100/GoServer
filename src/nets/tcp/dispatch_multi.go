// +build net_multi

package tcp

import (
	"common"
	"encoding/binary"
	"gamelog"
	"io"
	"nets/rpc"
	"time"
)

type netbuf struct { //替代bufio.Reader，引用同片内存
	rd  io.Reader
	buf []byte
	r   int
	w   int
}

func (b *netbuf) Init(r io.Reader)  { b.rd, b.buf = r, make([]byte, 4096) }
func (b *netbuf) Reset(r io.Reader) { b.rd, b.r, b.w = r, 0, 0 }
func (b *netbuf) ReadN(n int) (ret []byte) {
	if b.w-b.r >= n {
		return b.readBuf(n)
	}
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}
	for {
		if i, e := b.rd.Read(b.buf[b.w:]); e != nil || i < 0 {
			gamelog.Debug(e.Error())
			return nil
		} else if b.w += i; b.w-b.r >= n {
			return b.readBuf(n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func (b *netbuf) readBuf(n int) (ret []byte) {
	ret = b.buf[b.r : b.r+n]
	if b.r += n; b.r == b.w {
		b.r, b.w = 0, 0
	}
	return
}

func (self *TCPConn) readLoop() {
	var msgLen int
	var msgBuf []byte
	//msgBuf := make([]byte, Msg_Size_Max)
	var req, ack common.NetPack
	ack.Reset(make([]byte, common.PACK_HEADER_SIZE, 32), common.PACK_HEADER_SIZE)
	for q, r := &rpc.G_RpcQueue, &self.reader; ; {
		if self.IsClose() {
			break
		}
		//if _, err = io.ReadAtLeast(r, msgBuf[:kHeadLen], kHeadLen); err != nil {
		if msgBuf = r.ReadN(kHeadLen); msgBuf == nil {
			break
		}
		msgLen = int(binary.LittleEndian.Uint16(msgBuf))
		if msgLen < common.PACK_HEADER_SIZE || msgLen > Msg_Size_Max {
			gamelog.Error("invalid msgLen: %d", msgLen)
			break
		}
		//req.Reset(msgBuf[:msgLen], common.PACK_HEADER_SIZE)
		//if _, err = io.ReadAtLeast(r, req.Data(), msgLen); err != nil {
		if msgBuf = r.ReadN(msgLen); msgBuf == nil {
			break
		}
		req.Reset(msgBuf, common.PACK_HEADER_SIZE)
		//FIXME: 消息加密、验证有效性，pipeline模式优化
		q.Handle(self, &req, &ack) //io线程直接处理，须考虑竞态
	}
	self.Close()
	self.writer.StopWait()
}
