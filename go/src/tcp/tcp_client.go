package tcp

import (
	"common"
	"fmt"
	"net"
	"time"
)

type TCPClient struct { //作为client玩家数据的一个模块
	Addr            string
	PendingWriteNum int
	TcpConn         *TCPConn
	OnConnected     func(*TCPConn)
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
	packet := common.NewNetPack(32)
	packet.SetOpCode(G_MsgId_Regist)
	packet.WriteString(srcModule)
	packet.WriteInt32(int32(srcID))
	for {
		if client.connect() {
			if client.TcpConn != nil {
				if client.OnConnected != nil {
					client.OnConnected(client.TcpConn)
				}
				go client.TcpConn.writeRoutine()
				client.TcpConn.WriteMsg(packet)
				client.TcpConn.readRoutine() //goroutine会阻塞在这里
			}
		}
		time.Sleep(3 * time.Second)
	}
}
func (client *TCPClient) connect() bool {
	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		fmt.Printf("connect to %s error :%s \n", client.Addr, err.Error())
		return false
	}
	if conn == nil {
		return false
	}

	if client.TcpConn != nil {
		//断线重连的新连接标记得重置，否则tcpConn.readRoutine.readLoop会直接break
		client.TcpConn.ResetConn(conn)
	} else {
		client.TcpConn = newTCPConn(conn, client.PendingWriteNum, nil)
	}
	return true
}

func (client *TCPClient) Close() {
	client.TcpConn.Close()
	client.TcpConn = nil
}
