/***********************************************************************
* @ tcp连接
* @ brief
	1、Notice：
		http的消息处理，是另开goroutine调用的，所以函数中可阻塞
		tcp的消息处理，是在readRoutine中及时调用的，所以函数中不能有阻塞调用
		否则“该条连接”的读会被挂起，c++中的话，整个系统的处理线程都会阻塞掉

	2、server端目前是一条连接两个goroutine(readRoutine/writeRoutine)
		假设5k玩家，就有1w个goroutine，太多了

	3、msghandler可考虑设计成：不执行逻辑，仅将消息加入buf队列，由一个goroutine来处理
		不过那5k个readRoutine貌似省不了哇，感觉单独一个goroutine处理消息也不会有性能提升
		且增加了风险，若某条消息有阻塞调用，后面的就得等了

* @ author zhoumf
* @ date 2016-8-3
***********************************************************************/
package tcp

import (
	"bufio"
	"encoding/binary"
	"gamelog"
	"io"
	"net"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	G_MsgId_Disconnect = 7111
	G_MsgId_Regist     = 7112
)

var (
	G_HandlerMsgMap = map[uint16]func(*TCPConn, []byte){
		G_MsgId_Regist: DoRegistToSvr,
	}
)

type TCPConn struct { //登录时将TCPConn指针写入player中
	conn       net.Conn
	reader     *bufio.Reader //包装conn减少conn.Read的io次数，见【common\net.go】
	writeChan  chan []byte
	isClose    bool
	onNetClose func()
	Data       interface{}
}

func newTCPConn(conn net.Conn, pendingWriteNum int) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.ResetConn(conn)
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	return tcpConn
}

//isClose标记仅在ResetConn、Close中设置，其它地方只读
func (tcpConn *TCPConn) ResetConn(conn net.Conn) {
	tcpConn.conn = conn
	tcpConn.reader = bufio.NewReader(conn)
	tcpConn.isClose = false
}
func (tcpConn *TCPConn) Close() {
	if tcpConn.isClose {
		return
	}
	tcpConn.conn.Close()
	tcpConn.doWrite(nil) //触发writeRoutine结束
	tcpConn.isClose = true

	if tcpConn.onNetClose != nil {
		tcpConn.onNetClose()
	}
}

// msgdata must not be modified by other goroutines
func (tcpConn *TCPConn) WriteMsg(msgID uint16, msgdata []byte) {
	msgLen := uint16(len(msgdata))

	msgbuffer := make([]byte, 4+msgLen) //前2字节-msgLen；后2字节-msgID

	binary.LittleEndian.PutUint16(msgbuffer, msgLen)
	binary.LittleEndian.PutUint16(msgbuffer[2:], msgID)

	copy(msgbuffer[4:], msgdata)

	if false == tcpConn.isClose {
		tcpConn.doWrite(msgbuffer)
	}
}
func (tcpConn *TCPConn) doWrite(buf []byte) {
	select {
	case tcpConn.writeChan <- buf: //chan满后再写即阻塞，select进入default分支报错
	default:
		gamelog.Error("doWrite: channel full")
		tcpConn.conn.(*net.TCPConn).SetLinger(0)
		tcpConn.Close()
		// close(tcpConn.writeChan) //client重连chan里的数据得保留，server都是新new的
	}
}
func (tcpConn *TCPConn) writeRoutine() {
	for buf := range tcpConn.writeChan {
		if buf == nil {
			break
		}
		_, err := tcpConn.conn.Write(buf)
		if err != nil {
			gamelog.Error("WriteRoutine error: %s", err.Error())
			break
		}
	}
	tcpConn.Close()
}
func (tcpConn *TCPConn) readRoutine() {
	tcpConn.readLoop()
	tcpConn.Close()

	//通知业务层net断开
	tcpConn.msgDispatcher(G_MsgId_Disconnect, nil)
}
func (tcpConn *TCPConn) readLoop() error {
	var err error
	var msgHeader = make([]byte, 4) //前2字节-msgLen；后2字节-msgID
	var msgID uint16
	var msgLen uint16
	var firstTime bool = true

	for {
		if tcpConn.isClose {
			break
		}

		//TODO：client无需超时限制
		if firstTime == true {
			tcpConn.conn.SetReadDeadline(time.Now().Add(5000 * time.Second)) //首次读，5秒超时
			firstTime = false
		} else {
			tcpConn.conn.SetReadDeadline(time.Time{}) //后面读的就没有超时了
		}

		_, err = io.ReadFull(tcpConn.reader, msgHeader)
		if err != nil {
			gamelog.Error("ReadFull msgHeader error: %s", err.Error())
			return err
		}

		msgLen = binary.LittleEndian.Uint16(msgHeader)
		msgID = binary.LittleEndian.Uint16(msgHeader[2:])
		if msgLen <= 0 || msgLen > 10240 {
			gamelog.Error("ReadProcess Invalid msgLen :%d", msgLen)
			break
		}

		msgData := make([]byte, msgLen)
		_, err = io.ReadFull(tcpConn.reader, msgData)
		if err != nil {
			gamelog.Error("ReadFull msgData error: %s", err.Error())
			return err
		}

		tcpConn.msgDispatcher(msgID, msgData)
	}
	return nil
}
func (tcpConn *TCPConn) msgDispatcher(msgID uint16, pdata []byte) {
	// gamelog.Info("---msgID:%d, dataLen:%d", msgID, len(pdata))
	msghandler, ok := G_HandlerMsgMap[msgID]
	if !ok {
		gamelog.Error("msgid : %d have not a msg handler!!", msgID)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				gamelog.Error("msgID %d Error  %s", msgID, debug.Stack())
			}
		}
	}()
	msghandler(tcpConn, pdata)
}
