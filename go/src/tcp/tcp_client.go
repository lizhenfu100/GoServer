package tcp

import (
	"common"
	"gamelog"
	"net"
	"time"
)

type TCPClient struct { //作为client玩家数据的一个模块
	Addr            string
	PendingWriteNum int
	TcpConn         *TCPConn
}
type Msg_Regist_To_TcpSvr struct {
	Module string
	ID     int
}

func (client *TCPClient) ConnectToSvr(addr, srcModule string, srcID int) {
	client.Addr = addr
	client.PendingWriteNum = 32
	client.TcpConn = nil

	go client.connectRoutine(srcModule, srcID) //会断线后自动重连
}
func (client *TCPClient) connectRoutine(srcModule string, srcID int) {
	b, _ := common.ToBytes(&Msg_Regist_To_TcpSvr{srcModule, srcID})
	for {
		if client.connect() {
			if client.TcpConn != nil {
				go client.TcpConn.writeRoutine()
				client.TcpConn.WriteMsg(G_MsgId_Regist, b)
				client.TcpConn.readRoutine() //goroutine会阻塞在这里
			}
		}
		time.Sleep(3 * time.Second)
	}
}
func (client *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		gamelog.Error("connect to %s error :%s", client.Addr, err.Error())
		return false
	}
	if conn == nil {
		return false
	}

	if client.TcpConn != nil {
		client.TcpConn.ResetConn(conn)
		client.TcpConn.isClose = false //断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
	} else {
		client.TcpConn = newTCPConn(conn, client.PendingWriteNum)
	}
	return true
}

func (client *TCPClient) Close() {
	client.TcpConn.Close()
	client.TcpConn = nil
}
