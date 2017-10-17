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

	5、现在的架构是：每条连接各线程收数据，直接在io线程调注册的业务函数，对强交互的业务不友好
		要是做MMO之类的，可考虑像c++一样，io线程只负责收发，数据交付给全局队列，主线程逐帧处理，避免竞态

* @ reconnect
	1、Accept返回的conn，先收一个包，内含connId
	2、connId为0表示新连接，挑选一个空闲的TCPConn，newTCPConn()
	3、不为0即重连，取对应TCPConn，若它关闭的，随即ResetConn()

* @ 更稳定的连接
	*、参考项目
		【http://blog.codingnow.com/2014/02/connection_reuse.html】
		【https://github.com/funny/snet】
	*、新连接建立
		上行包：
			1、ConnID==0
			2、DH密钥
		下行包：
			1、加密的ConnID
			2、DH密钥
	*、连接修复(断线重连)
		上行包：
			1、旧有的ConnID
			2、Client已发送字节数
			3、Client已接收字节数
			4、密钥计算出的MD5
		服务器：
			1、验证合法性，失败立即断开
			2、根据ConnID定位旧连接，并下发“已发、已收字节数”作为重连回应
			3、再由Client上报的“已接收字节数”，计算出需重传数据，立即下发
			4、Client收到重连响应后，比较收发字节数差值来读取Server下发的重传数据

* @ author zhoumf
* @ date 2016-8-3
***********************************************************************/
package tcp

import (
	"bufio"
	"common"
	"encoding/binary"
	"errors"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	G_Msg_Size_Max = 1024
	Write_Chan_Cap = 32
)

var (
	G_HandleFunc = [enum.RpcEnumCnt]func(*common.NetPack, *common.NetPack, *TCPConn){
		enum.Rpc_regist:     DoRegistToSvr,
		enum.Rpc_svr_accept: OnSvrAcceptConn,
	}
	g_rpc_response = make(map[uint64]func(*common.NetPack))
	g_auto_req_idx = uint32(0)
	g_rw_lock      sync.RWMutex

	g_error_conn_close = errors.New("tcp conn close")
)

type TCPConn struct {
	conn       net.Conn
	reader     *bufio.Reader //包装conn减少conn.Read的io次数，见【common\net.go】
	writer     *bufio.Writer
	writeChan  chan []byte
	isClose    bool //isClose标记仅在ResetConn、Close中设置，其它地方只读
	onNetClose func(*TCPConn)
	UserPtr    interface{}
	sendBuffer *common.NetPack //for rpc
	backBuffer *common.NetPack
}

func newTCPConn(conn net.Conn, callback func(*TCPConn)) *TCPConn {
	self := new(TCPConn)
	self.ResetConn(conn)
	self.onNetClose = callback
	self.writeChan = make(chan []byte, Write_Chan_Cap)
	self.sendBuffer = common.NewNetPackCap(128)
	self.backBuffer = common.NewNetPackCap(128)
	return self
}
func (self *TCPConn) ResetConn(conn net.Conn) {
	self.conn = conn
	self.reader = bufio.NewReader(conn)
	self.writer = bufio.NewWriter(conn)
	self.isClose = false
}
func (self *TCPConn) Close() {
	if self.isClose {
		return
	}
	self.conn.Close()
	self.WriteBuf(nil) //触发writeRoutine结束
	self.isClose = true

	if self.onNetClose != nil {
		time.AfterFunc(30*time.Second, func() { //Notice:AfterFunc是在另一线程执行，所以调的函数须是线程安全的
			if self.isClose {
				self.onNetClose(self)
			}
		})
	}
}
func (self *TCPConn) IsClose() bool { return self.isClose }

// msgdata must not be modified by other goroutines
func (self *TCPConn) WriteMsg(msg *common.NetPack) {
	msgLen := uint16(msg.Size())

	//【Notice: chan里传递的是地址，这里不能像readLoop中那样，优化为"操作同一块buf"，必须每次new新的】
	//【否则writeRoutine里拿到的极可能是同样数据】
	buf := make([]byte, 2+msgLen)

	binary.LittleEndian.PutUint16(buf, msgLen)

	copy(buf[2:], msg.DataPtr)

	if false == self.isClose {
		self.WriteBuf(buf)
	}
}
func (self *TCPConn) WriteBuf(buf []byte) {
	select {
	case self.writeChan <- buf: //chan满后再写即阻塞，select进入default分支报错
	default:
		gamelog.Error("WriteBuf: channel full")
		self.conn.(*net.TCPConn).SetLinger(0)
		self.Close()
		// close(self.writeChan) //client重连chan里的数据得保留，server都是新new的
	}
}
func (self *TCPConn) _WriteFull(buf []byte) (err error) {
	if buf == nil {
		return g_error_conn_close
	}
	var n, nn int
	length := len(buf)
	for n < length && err == nil {
		nn, err = self.writer.Write(buf[n:]) //【Notice: WriteFull】bufio包装过，这里不会陷入系统调用；先缓存完chan的数据再Flush，更高效
		n += nn
	}
	if err != nil {
		gamelog.Error("WriteRoutine Write error: %s", err.Error())
	}
	return
}
func (self *TCPConn) writeRoutine() {
LOOP:
	for {
	goto_handle_chan:
		select {
		case buf := <-self.writeChan:
			if self._WriteFull(buf) != nil {
				break LOOP
			}
		default:
			var err error
			for i := 0; i < 100; i++ { //还写不完，等下一轮调度吧
				if err = self.writer.Flush(); err != io.ErrShortWrite {
					break
				}
			}
			if err != nil {
				gamelog.Error("WriteRoutine Flush error: %s", err.Error())
				break LOOP
			}
			//FIXME: 这里能不能加句 sleep(10ms)，让包更易积累，合并发送；不加的话，调度权就完全依赖golang了，其实只要不是写一个包就唤醒一次，问题不大
			//! block
			buf := <-self.writeChan
			if self._WriteFull(buf) != nil {
				break LOOP
			}
			goto goto_handle_chan
		}
	}
	self.Close()
}
func (self *TCPConn) readRoutine() {
	var err error
	var msgLen int
	var msgHeader = make([]byte, 2) //前2字节存msgLen
	var msgBuf = make([]byte, G_Msg_Size_Max)
	var packet = common.NewNetPack(nil)
	// var firstTime bool = true
	for {
		if self.isClose {
			break
		}
		//【check client heartbeat to close invalid conn】
		// if firstTime == true {
		// 	self.conn.SetReadDeadline(time.Now().Add(5 * time.Second)) //首次读，5秒超时【Notice: Client无需超时限制】
		// 	firstTime = false
		// } else {
		// 	self.conn.SetReadDeadline(time.Time{}) //后面读的就没有超时了
		// }
		_, err = io.ReadFull(self.reader, msgHeader)
		if err != nil {
			gamelog.Error("ReadFull msgHeader error: %s", err.Error())
			break
		}

		msgLen = int(binary.LittleEndian.Uint16(msgHeader))
		if msgLen <= 0 || msgLen > G_Msg_Size_Max {
			gamelog.Error("ReadProcess Invalid msgLen :%d", msgLen)
			break
		}
		packet.Reset(msgBuf[:msgLen])

		_, err = io.ReadFull(self.reader, packet.DataPtr)
		if err != nil {
			gamelog.Error("ReadFull msgData error: %s", err.Error())
			break
		}

		//FIXME: 消息加密、验证有效性，不通过即踢掉

		//【Notice: 目前是在io线程直接调消息响应函数(多线程处理玩家操作)，玩家之间互改数据须考虑竞态问题(可用actor模式解决)】
		//【Notice: 若友好支持玩家强交互，可将packet放入主逻辑循环的消息队列(SafeQueue)】
		self.msgDispatcher(packet)
	}
	self.Close()
}
func (self *TCPConn) msgDispatcher(msg *common.NetPack) {
	msgID := msg.GetOpCode()
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgID, r, debug.Stack())
		}
	}()
	if msghandler := G_HandleFunc[msgID]; msghandler != nil {
		gamelog.Debug("TcpMsgId:%d, len:%d", msgID, msg.Size())
		self.backBuffer.ResetHead(msg)
		msghandler(msg, self.backBuffer, self)
		if self.backBuffer.BodySize() > 0 {
			self.WriteMsg(self.backBuffer)
		}
	} else if rpcRecv := _FindResponse(msg.GetReqKey()); rpcRecv != nil {
		rpcRecv(msg)
		_DeleteResponse(msg.GetReqKey())
	} else {
		gamelog.Error("msgid[%d] havn't msg handler!!", msgID)
	}
}

// ------------------------------------------------------------
//! rpc 【非线程安全的】只给Player用
func (self *TCPConn) CallRpcUnsafe(msgID uint16, sendFun, recvFun func(*common.NetPack)) {
	if G_HandleFunc[msgID] != nil {
		gamelog.Error("Server and Client have the same Rpc[%d]", msgID)
		return
	}
	self.sendBuffer.SetOpCode(msgID)
	self.sendBuffer.SetReqIdx(_GetNextReqIdx())
	sendFun(self.sendBuffer)
	self.WriteMsg(self.sendBuffer)
	self.sendBuffer.Clear()

	if recvFun != nil {
		_InsertResponse(self.sendBuffer.GetReqKey(), recvFun)
	}
}
func (self *TCPConn) CallRpc(msgID uint16, sendFun, recvFun func(*common.NetPack)) {
	if G_HandleFunc[msgID] != nil {
		gamelog.Error("Server and Client have the same Rpc[%d]", msgID)
		return
	}
	buf := common.NewNetPackCap(128)
	buf.SetOpCode(msgID)
	buf.SetReqIdx(_GetNextReqIdx())
	sendFun(buf)
	self.WriteMsg(buf)

	if recvFun != nil {
		_InsertResponse(buf.GetReqKey(), recvFun)
	}
}
func _FindResponse(reqKey uint64) func(*common.NetPack) {
	g_rw_lock.RLock()
	defer g_rw_lock.RUnlock()
	if rpcRecv, ok := g_rpc_response[reqKey]; ok {
		return rpcRecv
	}
	return nil
}
func _InsertResponse(reqKey uint64, fun func(*common.NetPack)) {
	// 后来的应该覆盖之前的
	// if _FindResponse(reqKey) == nil {
	g_rw_lock.Lock()
	g_rpc_response[reqKey] = fun
	g_rw_lock.Unlock()
	// }
}
func _DeleteResponse(reqKey uint64) {
	g_rw_lock.Lock()
	delete(g_rpc_response, reqKey)
	g_rw_lock.Unlock()
}
func _GetNextReqIdx() uint32 { return atomic.AddUint32(&g_auto_req_idx, 1) }
