/***********************************************************************
* @ 与玩家强绑定的rpc，比对net_rpc.go
* @ brief
	1、将原生rpc的参数转换为 player *TPlayer

	2、拦截原生网络rpc，处理通用部分得到*TPlayer，再转入PlayerRpc

* @ author zhoumf
* @ date 2018-3-23
***********************************************************************/
package player

import (
	"common"
	"common/std/compress"
	"gamelog"
	"generate_out/rpc/enum"
	"net/http"
	"netConfig"
	http2 "nets/http/http"
	"nets/tcp"
	"sync/atomic"
)

type PlayerRpc func(req, ack *common.NetPack, this *TPlayer)

var G_PlayerHandleFunc [enum.RpcEnumCnt]PlayerRpc

// 访问玩家数据的消息，要求该玩家已在缓存中，否则不处理
//【Notice：登录、创建角色，可做成普通rpc，用以建立玩家缓存】
func RegPlayerRpc(list map[uint16]PlayerRpc) {
	for k, v := range list {
		G_PlayerHandleFunc[k] = v
	}
	tcp.RegHandlePlayerRpc(_PlayerRpcTcp)    //tcp 直连
	http2.RegHandlePlayerRpc(_PlayerRpcHttp) //http 直连
}
func DoPlayerRpc(this *TPlayer, rpcId uint16, req, ack *common.NetPack) bool {
	if msgFunc := G_PlayerHandleFunc[rpcId]; msgFunc != nil {
		atomic.StoreUint32(&this._idleMin, 0)
		msgFunc(req, ack, this)
		return true
	}
	return false
}

// ------------------------------------------------------------
// - 将原生tcpRpc的 "conn *tcp.TCPConn" 参数转换为 "player *TPlayer"
func _PlayerRpcTcp(req, ack *common.NetPack, conn *tcp.TCPConn) bool {
	rpcId := req.GetOpCode()
	if player, ok := conn.UserPtr.(*TPlayer); ok {
		DoPlayerRpc(player, rpcId, req, ack)
	}
	return G_PlayerHandleFunc[rpcId] != nil
}
func _PlayerRpcHttp(w http.ResponseWriter, r *http.Request) {
	buf := http2.ReadRequest(r) //! 接收信息
	if buf == nil {
		return
	}
	req := common.NewNetPack(buf)
	if req == nil {
		gamelog.Error("invalid req: %v", buf)
		return
	}
	msgId := req.GetOpCode()
	if msgId >= enum.RpcEnumCnt {
		gamelog.Error("PlayerMsg(%d) Not Regist", msgId)
		req.Free()
		return
	}
	if msgId != enum.Rpc_game_heart_beat {
		gamelog.Debug("HttpMsg:%d, len:%d", msgId, req.Size())
	}
	//defer func() {//库已经有recover了，见net/http/server.go:1918
	//	if r := recover(); r != nil {
	//		gamelog.Error("recover msgId:%d\n%v: %s", msgId, r, debug.Stack())
	//	}
	//	ack.Free()
	//}()
	ack := common.NewNetPackCap(128) //! 创建回复
	accountId := req.GetReqIdx()
	//FIXME: http通信中途安全性不够，能修改client net pack里的uid，进而操作别人数据
	//FIXME: 账号服登录验证后下发给client的token，client应该保留，附在每个HttpReq里，防止恶意窜改他人数据
	if player := BeforeRecvHttpMsg(accountId); player != nil {
		if DoPlayerRpc(player, msgId, req, ack) {
			AfterRecvHttpMsg(player, ack)
		} else {
			gamelog.Error("PlayerMsg(%d) Not Regist", msgId)
		}
	} else {
		gamelog.Debug("Player(%d) isn't online", accountId)
	}

	compress.CompressTo(ack.Data(), w)
	ack.Free()
	req.Free()
}

// ------------------------------------------------------------
// - 网关转发的玩家消息
func Rpc_recv_player_msg(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()
	accountId := req.ReadUInt32()

	gamelog.Debug("PlayerMsg:%d", rpcId)

	if player := FindAccountId(accountId); player != nil {
		if !DoPlayerRpc(player, rpcId, req, ack) {
			gamelog.Error("PlayerMsg(%d) Not Regist", rpcId)
		}
	} else {
		gamelog.Debug("Player(%d) isn't online", accountId)
	}
}

// ------------------------------------------------------------
// - 与其它玩家交互(可能位于其它节点，能通知到别人客户端)
func CallRpcPlayer(accountId uint32, msgId uint16, sendFun, recvFun func(*common.NetPack)) {
	if msgFunc := G_PlayerHandleFunc[msgId]; msgFunc != nil {
		if player := FindAccountId(accountId); player != nil {
			req := common.NewNetPackCap(64)
			ack := common.NewNetPackCap(64)
			req.SetOpCode(msgId)
			sendFun(req)
			msgFunc(req, ack, player)
			if recvFun != nil {
				recvFun(ack)
			}
			req.Free()
			ack.Free()
			return
		}
	}
	netConfig.CallRpcGateway(accountId, msgId, sendFun, recvFun)
}
