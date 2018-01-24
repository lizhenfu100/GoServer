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
	3、不为0即重连，取对应TCPConn，若它关闭的，随即resetConn()

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
	"sync/atomic"
	"time"
)

const (
	G_Msg_Size_Max    = 1024
	Write_Chan_Cap    = 32
	Delay_Delete_Conn = 30 * time.Second
)

var (
	G_HandleFunc = [enum.RpcEnumCnt]func(*common.NetPack, *common.NetPack, *TCPConn){
		enum.Rpc_regist:           DoRegistToSvr,
		enum.Rpc_svr_accept:       OnSvrAcceptConn,
		enum.Rpc_report_net_error: ReportNetError,
	}
	G_RpcQueue = NewRpcQueue(4096)
)

type TCPConn struct {
	conn          net.Conn
	reader        *bufio.Reader //包装conn减少conn.Read的io次数，见【common\net.go】
	writer        *bufio.Writer
	writeChan     chan []byte
	_isClose      int32 //isClose标记仅在resetConn、Close中设置，其它地方只读
	_isWriteClose bool
	onNetClose    func(*TCPConn)
	delayDel      *time.Timer //延时清理连接，提高重连效率
	UserPtr       interface{}
}

func newTCPConn(conn net.Conn) *TCPConn {
	self := new(TCPConn)
	self.writeChan = make(chan []byte, Write_Chan_Cap)
	self.resetConn(conn)
	return self
}
func (self *TCPConn) resetConn(conn net.Conn) {
	self.conn = conn
	self.reader = bufio.NewReader(conn)
	self.writer = bufio.NewWriter(conn)
	atomic.StoreInt32(&self._isClose, 0)
}
func (self *TCPConn) Close() {
	if self.IsClose() {
		return
	}
	self.conn.(*net.TCPConn).SetLinger(0) //丢弃数据
	self.conn.Close()
	if !self._isWriteClose {
		self.WriteBuf(nil) //触发writeRoutine结束
	}
	atomic.StoreInt32(&self._isClose, 1)

	//通知逻辑线程，连接断开
	packet := common.NewNetPackCap(16)
	packet.SetOpCode(enum.Rpc_report_net_error)
	G_RpcQueue.Insert(self, packet)
}
func (self *TCPConn) IsClose() bool { return atomic.LoadInt32(&self._isClose) > 0 }

// msgdata must not be modified by other goroutines
func (self *TCPConn) WriteMsg(msg *common.NetPack) {
	msgLen := uint16(msg.Size())

	//【Notice: chan里传递的是地址，这里不能像readLoop中那样，优化为"操作同一块buf"，必须每次new新的】
	//【否则writeRoutine里拿到的极可能是同样数据】
	buf := make([]byte, 2+msgLen)

	binary.LittleEndian.PutUint16(buf, msgLen)

	copy(buf[2:], msg.Data())

	self.WriteBuf(buf)
}
func (self *TCPConn) WriteBuf(buf []byte) {
	select {
	case self.writeChan <- buf: //chan满后再写即阻塞，select进入default分支报错
	default:
		gamelog.Error("WriteBuf: channel full")
		self.Close()
		// close(self.writeChan) //client重连chan里的数据得保留，server都是新new的
	}
}
func (self *TCPConn) _WriteFull(buf []byte) (err error) { //FIXME：err可能是io.ErrShortWrite，网络还是能继续工作的
	if buf == nil || self.IsClose() {
		return errors.New("tcp conn close")
	}
	var n, nn int
	length := len(buf)
	for n < length && err == nil {
		nn, err = self.writer.Write(buf[n:]) //Notice：bufio包装过，这里不会陷入系统调用；先缓存完chan的数据再Flush，更高效
		n += nn
	}
	if err != nil {
		gamelog.Error("WriteRoutine Write error: %s", err.Error())
	}
	return
}
func (self *TCPConn) writeRoutine() {
	self._isWriteClose = false
LOOP:
	for {
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
		}
	}
	self._isWriteClose = true
	self.Close()
}
func (self *TCPConn) readRoutine() {
	var err error
	var msgLen int
	var msgHeader = make([]byte, 2) //前2字节存msgLen
	//var packet, msgBuf = &common.NetPack{}, make([]byte, G_Msg_Size_Max)
	for {
		if self.IsClose() {
			break
		}
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
		//packet.Reset(msgBuf[:msgLen]) //每次都操作的同片内存
		packet := common.NewNetPackLen(msgLen)

		_, err = io.ReadFull(self.reader, packet.Data())
		if err != nil {
			gamelog.Error("ReadFull msgData error: %s", err.Error())
			break
		}

		//FIXME: 消息加密、验证有效性，不通过即踢掉

		//【Notice: 在io线程直接调消息响应函数(多线程处理玩家操作)，玩家之间互改数据须考虑竞态问题(可用actor模式解决)】
		//【Notice: 若友好支持玩家强交互，可将packet放入主逻辑循环的消息队列(SafeQueue)】
		G_RpcQueue.Insert(self, packet)
	}
	self.Close()
}

// ------------------------------------------------------------
// rpc
func (self *TCPConn) CallRpc(msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	G_RpcQueue.CallRpc(self, msgId, sendFun, recvFun)
}

func ReportNetError(req, ack *common.NetPack, conn *TCPConn) {
	//if conn.onNetClose != nil { //迁移至：延时删除
	//	conn.onNetClose(conn)
	//}
}
