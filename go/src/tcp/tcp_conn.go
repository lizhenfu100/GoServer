package tcp

import (
	"encoding/binary"
	"gamelog"
	"io"
	"net"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	G_MSG_BEGIN      int16 = 0
	G_MSG_END        int16 = 1000
	G_MSG_DISCONNECT int16 = 111
)

type TCPConn struct { //登录时将TCPConn指针写入player中
	conn             net.Conn
	writeChan        chan []byte
	isClose          bool
	isCleaned        bool
	Data             interface{}
	onReadRoutineEnd func()
}

func newTCPConn(conn net.Conn, pendingWriteNum int) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	return tcpConn
}
func (tcpConn *TCPConn) close() {
	if tcpConn.isClose {
		return
	}
	tcpConn.conn.Close()
	tcpConn.doWrite(nil)
}

// msgdata must not be modified by other goroutines
func (tcpConn *TCPConn) WriteMsg(msgID int16, msgdata []byte) {
	msgLen := len(msgdata)

	msgbuffer := make([]byte, 6+msgLen) //前4字节-msgLen；后2字节-msgID

	binary.LittleEndian.PutUint32(msgbuffer, uint32(msgLen))

	binary.LittleEndian.PutUint16(msgbuffer[4:], uint16(msgID))

	copy(msgbuffer[6:], msgdata)

	if false == tcpConn.isClose {
		tcpConn.doWrite(msgbuffer)
	}
}
func (tcpConn *TCPConn) doWrite(buf []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		gamelog.Error("doWrite: channel full")
		tcpConn.doDestroy()
		return
	}
	tcpConn.writeChan <- buf
}
func (tcpConn *TCPConn) writeRoutine() {
	for buf := range tcpConn.writeChan {
		if buf == nil || tcpConn.isClose {
			break
		}
		_, err := tcpConn.conn.Write(buf)
		if err != nil {
			gamelog.Error("WriteRoutine error: %s", err.Error())
			break
		}
	}
	gamelog.Info("WriteRoutine Over !!!")
	tcpConn.conn.Close()
	tcpConn.isClose = true
}
func (tcpConn *TCPConn) doDestroy() {
	tcpConn.conn.(*net.TCPConn).SetLinger(0)
	tcpConn.conn.Close()
	close(tcpConn.writeChan)
	tcpConn.isClose = true
}

func (tcpConn *TCPConn) readRoutine() {
	tcpConn.readLoop()
	tcpConn.close()

	//通知client断开
	tcpConn.msgDispatcher(G_MSG_DISCONNECT, nil)

	if tcpConn.onReadRoutineEnd != nil {
		tcpConn.onReadRoutineEnd()
	}
}
func (tcpConn *TCPConn) readLoop() error {
	defer func() {
		tcpConn.conn.Close()
	}()

	var err error
	var msgHeader = make([]byte, 6)
	var msgID int16
	var msgLen int32
	var firstTime bool = true

	for {
		if tcpConn.isClose {
			break
		}

		if firstTime == true {
			tcpConn.conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //首次读，5秒超时
			firstTime = false
		} else {
			tcpConn.conn.SetReadDeadline(time.Time{}) //后面读的就没有超时了
		}

		_, err = io.ReadAtLeast(tcpConn.conn, msgHeader, 6) //前4字节-msgLen；后2字节-msgID
		if err != nil {
			gamelog.Error("ReadAtLeast msgHeader error: %s", err.Error())
			return err
		}

		msgLen = int32(binary.LittleEndian.Uint16(msgHeader[:4]))
		msgID = int16(binary.LittleEndian.Uint16(msgHeader[4:]))
		if msgLen <= 0 || msgLen > 10240 {
			gamelog.Error("ReadProcess Invalid msgLen :%d", msgLen)
			break
		}
		if msgID <= G_MSG_BEGIN || msgID >= G_MSG_END {
			gamelog.Error("ReadProcess Invalid msgID :%d", msgID)
			break
		}

		msgData := make([]byte, msgLen)
		_, err = io.ReadAtLeast(tcpConn.conn, msgData, int(msgLen))
		if err != nil {
			gamelog.Error("ReadAtLeast msgData error: %s", err.Error())
			return err
		}

		tcpConn.msgDispatcher(msgID, msgData)
	}
	return nil
}
func (tcpConn *TCPConn) msgDispatcher(msgID int16, pdata []byte) {
	// gamelog.Info("---msgID:%d, data:%d", msgID, pdata)
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
