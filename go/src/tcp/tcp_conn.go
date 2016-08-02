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

var (
	G_MSG_DISCONNECT int16 = 111
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
	tcpConn.conn = conn
	tcpConn.reader = bufio.NewReader(conn)
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	return tcpConn
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
func (tcpConn *TCPConn) WriteMsg(msgID int16, msgdata []byte) {
	msgLen := len(msgdata)

	msgbuffer := make([]byte, 4+msgLen) //前2字节-msgLen；后2字节-msgID

	binary.LittleEndian.PutUint16(msgbuffer, uint16(msgLen))
	binary.LittleEndian.PutUint16(msgbuffer[2:], uint16(msgID))

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
	tcpConn.msgDispatcher(G_MSG_DISCONNECT, nil)
}
func (tcpConn *TCPConn) readLoop() error {
	var err error
	var msgHeader = make([]byte, 4) //前2字节-msgLen；后2字节-msgID
	var msgID int16
	var msgLen int32
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

		msgLen = int32(binary.LittleEndian.Uint16(msgHeader))
		msgID = int16(binary.LittleEndian.Uint16(msgHeader[2:]))
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
func (tcpConn *TCPConn) msgDispatcher(msgID int16, pdata []byte) {
	// gamelog.Info("---msgID:%d, dataLen:%d", msgID, len(pdata))
	msghandler, ok := G_HandlerMap[msgID]
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
