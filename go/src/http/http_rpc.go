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
	"gamelog"
	"generate_out/rpc/enum"
	"net/http"
	"runtime/debug"
)

const (
	Client_ReLogin_Flag = 0xFFFFFFFF
)

var (
	G_HandleFunc       [enum.Rpc_enum_cnt]func(req, ack *common.NetPack)
	G_PlayerHandleFunc [enum.Rpc_enum_cnt]func(req, ack *common.NetPack, p interface{})

	//! 需要主动发给玩家的数据，每回通信时捎带过去
	G_Before_Recv_Player func(uint32) interface{}
	G_After_Recv_Player  func(interface{}, *common.NetPack)
)

//////////////////////////////////////////////////////////////////////
//! system rpc
func CallRpc(addr string, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	buf := common.NewNetPackCap(64)
	buf.SetOpCode(rid)
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
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
		}
	}()

	if handler := G_HandleFunc[msgId]; handler != nil {
		handler(req, ack)
		common.CompressInto(ack.DataPtr, w)
	} else {
		println("\n===> Http HandleRpc:", msgId, "Not Regist!!!")
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
func (self *PlayerRpc) CallRpc(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	buf := common.NewNetPackCap(64)
	buf.SetOpCode(rid)
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
func _RecvHttpSvrData(buf *common.NetPack) {
	//对应于 http_to_client.go
}
func RegHandlePlayerRpc() { http.HandleFunc("/player_rpc", _HandlePlayerRpc) }
func _HandlePlayerRpc(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewNetPackLen(int(r.ContentLength) - common.PACK_HEADER_SIZE)
	r.Body.Read(req.DataPtr)

	//! 创建回复
	ack := common.NewNetPackCap(128)
	msgId := req.GetOpCode()
	pid := req.GetReqIdx()
	defer func() {
		if r := recover(); r != nil {
			gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
		}
	}()
	//FIXME: 验证消息安全性，防改包
	//FIXME: http通信中途安全性不够，能修改client net pack里的pid，进而操作别人数据
	//FIXME: 账号服登录验证后下发给client的token，client应该保留，附在每个HttpReq里，防止恶意窜改他人数据

	if msgId != enum.Rpc_game_heart_beat {
		gamelog.Debug("HttpMsg:%d, len:%d, playerId:%d", msgId, req.Size(), pid)
	}
	if handler := G_PlayerHandleFunc[msgId]; handler != nil {
		var player interface{}
		if G_Before_Recv_Player != nil {
			player = G_Before_Recv_Player(pid)
		}
		if player == nil {
			gamelog.Debug("===> pid:%d isn't in memcache, please relogin", pid)
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
		println("\n===> Http HandlePlayerRpc:", msgId, "Not Regist!!!")
	}
}
