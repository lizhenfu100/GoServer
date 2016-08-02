/***********************************************************************
* @ 网络IO
* @ brief
	1、"io.ReadFull/ReadAtLeast"内部是循环调用"io.Read"的过程，潜在的IO次数很多
		因"io.Read"读到的数据可能小于预期长度
		在一个循环里调用`Read`，累加每次返回的`n`并对buf指针偏移再做下次`Read`
		直到`n`的累加值达到我们的预期

	2、"bufio.Reader"的基本工作原理是使用一块预先分配好的内存作为缓冲区，发生真实IO的时候尽量填充缓冲区
		调用者读取数据的时候先从缓冲区中读取，从而减少真实的IO调用次数，以起到优化作用

		(1)bufio内部是一个固定大内存(默认4096)，尽量一次读完所有数据，供上层逻辑解析完整包用
		(2)整个文件调底层Read的只有两个地方：
			*、一个是外部传入的[]byte超过4096，直接调底层Read，一次读完
			*、另一个是内部缓存已被读完，调了fill()，里面会用剩余的缓冲去读“b.rd.Read(b.buf[b.w:])”也是足够大的
		(3)效果上就是固定大内存去一次性尽可能读全部数据

	3、`io.Writer`在每次写数据时，会保证数据的完整写入，这个特性跟`io.Reader`正好是相反的
		基于`io.Writer`的这一特性，我们可以推断，当我们往一个`net.Conn`写入数据时，会阻塞
		可用chan将`io.Writer`变成异步行为

* @ author zhoumf
* @ date 2016-7-20
************************************************************************/
package common

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type PacketConn struct {
	net.Conn
	reader *bufio.Reader
}

func NewPacketConn(conn net.Conn) *PacketConn {
	return &PacketConn{conn, bufio.NewReader(conn)}
}
func (self *PacketConn) ReadPacket() ([]byte, error) { //用户取网络包接口
	//1、先读2字节头，里面记录了消息长度
	head := make([]byte, 2)
	if _, err := io.ReadFull(self.reader, head); err != nil {
		return nil, err
	}
	//2、解析出消息体长度，大端格式
	size := binary.BigEndian.Uint16(head)
	packet := make([]byte, size)

	//3、读出消息体
	if _, err := io.ReadFull(self.reader, packet); err != nil {
		return nil, err
	}
	return packet, nil
}

//! 异步`io.Writer`
type Session struct {
	conn     net.Conn
	sendChan chan []byte
}

func (self *Session) sendLoop() {
	// sendChan无数据，阻塞
	// conn.Write写完全部数据之前，阻塞
	for {
		buf := <-self.sendChan
		if _, err := self.conn.Write(buf); err != nil {
			return // 出现IO失败就停止
		}
	}
}
func (self *Session) Send(buf []byte) error {
	// sendChan被写满，阻塞，select进入default分支报错
	select {
	case self.sendChan <- buf:
	default:
		return errors.New("Send Chan Blocked!")
	}
	return nil
}
func NewSession(conn net.Conn, sendChanSize int) *Session {
	s := &Session{conn, make(chan []byte, sendChanSize)}
	go s.sendLoop()
	return s
}
