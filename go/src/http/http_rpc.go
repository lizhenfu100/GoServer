/***********************************************************************
* @ http rpc
* @ brief
	1、system rpc：将原生http的参数统一转换为NetPack
	2、player rpc：在system rpc基础之上，加了层find player逻辑，若找不到不处理

* @ author zhoumf
* @ date 2017-8-10
***********************************************************************/
package http

import (
	"common"
	"encoding/binary"
	//"fmt"
	"net/http"
)

const (
	Client_ReLogin_Flag = 0xFFFFFFFF
)

var (
	G_HandlerMap       = make(map[uint16]func(req, ack *common.NetPack))
	G_PlayerHandlerMap = make(map[uint16]func(req, ack *common.NetPack, p interface{}))

	//! 需要主动发给玩家的数据，每回通信时捎带过去
	G_Before_Recv_Player func(uint32) interface{}
	G_After_Recv_Player  func(interface{}, *common.NetPack)
)

//////////////////////////////////////////////////////////////////////
//! system rpc
func CallRpc(addr string, rpc string, sendFun, recvFun func(*common.NetPack)) {
	buf := common.NewNetPackCap(64)
	buf.SetRpc(rpc)
	sendFun(buf)
	b := PostReq(addr+"client_rpc", buf.DataPtr)
	if recvFun != nil {
		b2 := common.Decompress(b)
		recvFun(common.NewNetPack(b2))
	}
}
func RegHandleRpc() { http.HandleFunc("/client_rpc", _HandleRpc) }
func _HandleRpc(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength) - common.PACK_HEADER_SIZE)
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(128)

	msgId := req.GetOpCode()

	if handler, ok := G_HandlerMap[msgId]; ok {

		handler(req, ack)

		common.CompressInto(ack.DataPtr, w)
	} else {
		println("\n===> Http HandleRpc:", common.DebugRpcIdToName(msgId), "Not Regist!!!")
		//fmt.Println(req)
	}
}

//////////////////////////////////////////////////////////////////////
//! player rpc
type PlayerRpc struct {
	Url      string
	PlayerId uint32
}

func NewPlayerRpc(addr string, pid uint32) *PlayerRpc {
	return &PlayerRpc{addr + "player_rpc", pid}
}
func (self *PlayerRpc) CallRpc(rpc string, sendFun, recvFun func(*common.NetPack)) {
	buf := common.NewNetPackCap(64)
	buf.SetRpc(rpc)
	buf.SetReqIdx(self.PlayerId)
	sendFun(buf)
	b := PostReq(self.Url, buf.DataPtr)
	b2 := common.Decompress(b)
	recvBuf := common.NewNetPack(b2)
	if recvFun != nil {
		recvFun(recvBuf)
	}
	_RecvHttpSvrData(recvBuf) //服务器主动下发的数据
}
func RegHandlePlayerRpc() { http.HandleFunc("/player_rpc", _HandlePlayerRpc) }
func _HandlePlayerRpc(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength) - common.PACK_HEADER_SIZE)
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(128)

	//FIXME: 验证消息安全性，防改包
	//FIXME: http通信中途安全性不够，能修改client net pack里的pid，进而操作别人数据
	//FIXME: 账号服登录验证后下发给client的token，client应该保留，附在每个HttpReq里，防止恶意窜改他人数据

	msgId := req.GetOpCode()
	pid := req.GetReqIdx()

	//心跳包太碍眼了
	if msgId != 10026 {
		println("\nHttpMsg:", common.DebugRpcIdToName(msgId), "len:", req.Size(), "  playerId:", pid)
	}
	if handler, ok := G_PlayerHandlerMap[msgId]; ok {
		var player interface{}
		if G_Before_Recv_Player != nil {
			player = G_Before_Recv_Player(pid)
		}
		if player == nil {
			flag := make([]byte, 4) //重登录标记
			binary.LittleEndian.PutUint32(flag, Client_ReLogin_Flag)
			w.Write(flag)
			return
		}

		handler(req, ack, player)

		if G_After_Recv_Player != nil && player != nil {
			G_After_Recv_Player(player, ack)
		}

		common.CompressInto(ack.DataPtr, w)

	} else {
		println("\n===> Http HandlePlayerRpc:", common.DebugRpcIdToName(msgId), "Not Regist!!!")
	}
}
func _RecvHttpSvrData(buf *common.NetPack) {
	//对应于 http_to_client.go
}
