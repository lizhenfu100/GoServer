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

	4、Rpc:
		g_rpc_response须加读写锁，与c++(多线程收-主线程顺序处理)不同，go是每个用户一条goroutine

* @ author zhoumf
* @ date 2016-8-3
***********************************************************************/
package tcp

import (
	"bufio"
	"common"
	"encoding/binary"
	"gamelog"
	"io"
	"net"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

const (
	G_Msg_Size_Max = 1024
	G_MsgId_Regist = 60000
)

var (
	G_HandlerMsgMap = map[uint16]func(*TCPConn, *common.NetPack){
		G_MsgId_Regist: DoRegistToSvr,
	}
	g_rpc_response = make(map[uint16]func(*common.NetPack))
	g_rw_lock      = new(sync.RWMutex)
)

type TCPConn struct {
	conn       net.Conn
	reader     *bufio.Reader //包装conn减少conn.Read的io次数，见【common\net.go】
	writeChan  chan []byte
	isClose    bool //isClose标记仅在ResetConn、Close中设置，其它地方只读
	onNetClose func(*TCPConn)
	UserPtr    interface{}
	sendBuffer *common.NetPack
	BackBuffer *common.NetPack
}

func newTCPConn(conn net.Conn, pendingWriteNum int, callback func(*TCPConn)) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.ResetConn(conn)
	tcpConn.onNetClose = callback
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.sendBuffer = common.NewNetPackCap(128)
	tcpConn.BackBuffer = common.NewNetPackCap(128)
	return tcpConn
}

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
		tcpConn.onNetClose(tcpConn)
	}
}

// msgdata must not be modified by other goroutines
func (tcpConn *TCPConn) WriteMsg(msg *common.NetPack) {
	msgLen := uint16(msg.Size())

	buf := make([]byte, 2+msgLen)

	binary.LittleEndian.PutUint16(buf, msgLen)

	copy(buf[2:], msg.DataPtr)

	if false == tcpConn.isClose {
		tcpConn.doWrite(buf)
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
}
func (tcpConn *TCPConn) readLoop() error {
	var err error
	var msgLen int
	var msgHeader = make([]byte, 2) //前2字节-msgLen
	var msgBuf = make([]byte, G_Msg_Size_Max)
	var packet = common.NewNetPack(nil)
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

		msgLen = int(binary.LittleEndian.Uint16(msgHeader))
		if msgLen <= 0 || msgLen > G_Msg_Size_Max {
			gamelog.Error("ReadProcess Invalid msgLen :%d", msgLen)
			break
		}

		packet.Reset(msgBuf[:msgLen])

		_, err = io.ReadFull(tcpConn.reader, packet.DataPtr)
		if err != nil {
			gamelog.Error("ReadFull msgData error: %s", err.Error())
			return err
		}

		tcpConn.msgDispatcher(packet.GetOpCode(), packet)
	}
	return nil
}
func (tcpConn *TCPConn) msgDispatcher(msgID uint16, msg *common.NetPack) {
	// gamelog.Info("---msgID:%d, dataLen:%d", msgID, msg.Size())
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				gamelog.Error("recover msgID:%d %s", msgID, debug.Stack())
			}
		}
	}()
	if msghandler, ok := G_HandlerMsgMap[msgID]; ok {
		tcpConn.BackBuffer.ResetHead(msg)
		msghandler(tcpConn, msg)
		if tcpConn.BackBuffer.BodySize() > 0 {
			tcpConn.WriteMsg(tcpConn.BackBuffer)
		}
	} else if rpcRecv := _FindResponse(msgID); rpcRecv != nil {
		rpcRecv(msg)
	} else {
		gamelog.Error("msgid[%d] havn't msg handler!!", msgID)
	}
}

//! rpc
func (tcpConn *TCPConn) CallRpc(rpc string, sendFun func(*common.NetPack)) uint16 {
	msgID := common.RpcNameToId(rpc)
	tcpConn.sendBuffer.ClearBody()
	tcpConn.sendBuffer.SetOpCode(msgID)
	sendFun(tcpConn.sendBuffer)
	tcpConn.WriteMsg(tcpConn.sendBuffer)
	return msgID
}
func (tcpConn *TCPConn) CallRpc2(rpc string, sendFun, recvFun func(*common.NetPack)) {
	msgID := tcpConn.CallRpc(rpc, sendFun)
	// insert recvFun
	if _, ok := G_HandlerMsgMap[msgID]; ok {
		gamelog.Error("Server and Client have the same Rpc[%s]", rpc)
	} else if rpcRecv := _FindResponse(msgID); rpcRecv == nil {
		g_rpc_response[msgID] = recvFun
	}
}
func _FindResponse(msgid uint16) func(*common.NetPack) {
	g_rw_lock.RLock()
	defer g_rw_lock.RUnlock()
	if rpcRecv, ok := g_rpc_response[msgid]; ok {
		return rpcRecv
	}
	return nil
}
func _InsertResponse(msgid uint16, fun func(*common.NetPack)) {
	if _FindResponse(msgid) == nil {
		g_rw_lock.Lock()
		g_rpc_response[msgid] = fun
		g_rw_lock.Unlock()
	}
}
