package logic

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/tcp"
)

func Rpc_gateway_relay_player_msg(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()     //目标rpc
	accountId := req.ReadUInt32() //目标玩家 accountId := _GetAidOfConn(req, conn)
	oldReqKey := req.GetReqKey()
	gamelog.Debug("relay_player_msg(%d)", rpcId)

	if accountId == 0 {
		gamelog.Debug("accountId nil")
		return
	} else if netConfig.HashGatewayID(accountId) == meta.G_Local.SvrID { //应连本节点的玩家
		RelayPlayerMsg(accountId, rpcId, req.LeftBuf(), oldReqKey, conn)
	} else {
		//非本节点玩家，转至其它gateway
		netConfig.CallRpcGateway(accountId, rpcId, func(buf *common.NetPack) {
			buf.WriteBuf(req.LeftBuf())
		}, func(backBuf *common.NetPack) {
			//异步回调，不能直接用ack
			backBuf.SetReqKey(oldReqKey)
			conn.WriteMsg(backBuf)
		})
	}
}
func RelayPlayerMsg(accountId uint32, rpcId uint16, rpcData []byte, oldReqKey uint64, conn *tcp.TCPConn) {
	rpcModule := enum.GetRpcModule(rpcId)
	if rpcModule == "" {
		gamelog.Error("rpc(%d) havn't module", rpcId)
		return
	}
	if rpcModule == "client" { //转给客户端
		if pConn := GetClientConn(accountId); pConn != nil {
			pConn.CallRpc(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(rpcData)
			}, func(backBuf *common.NetPack) {
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		} else {
			gamelog.Debug("rid(%d) accountId(%d) client conn nil", rpcId, accountId)
		}
	} else { //转给后台节点
		if pConn := _GetSvrNodeConn(rpcModule, accountId); pConn != nil {
			pConn.CallRpc(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
				buf.WriteUInt16(rpcId)
				buf.WriteUInt32(accountId)
				buf.WriteBuf(rpcData)
			}, func(backBuf *common.NetPack) {
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		} else {
			gamelog.Debug("rid(%d) accountId(%d) svr conn nil", rpcId, accountId)
		}
	}
}

// ------------------------------------------------------------
// 辅助函数
func _GetAidOfConn(req *common.NetPack, conn *tcp.TCPConn) uint32 {
	if accountId, ok := conn.UserPtr.(uint32); ok { //客户端的连接，绑定了aid，不必消息附带(还更安全)
		return accountId
	} else if _, ok := conn.UserPtr.(*meta.Meta); ok { //后台节点之间的消息，均带有aid，用以定位玩家
		accountId := req.ReadUInt32()
		return accountId
	}
	return 0
}
func _GetSvrNodeConn(module string, aid uint32) *tcp.TCPConn {
	if module == "game" {
		return GetGameConn(aid)
	} else {
		// 其它节点无状态的，AccountId取模得节点id
		ids := meta.GetModuleIDs(module, meta.G_Local.Version)
		if length := uint32(len(ids)); length > 0 {
			id := ids[aid%length]
			return netConfig.GetTcpConn(module, id)
		}
	}
	return nil
}

// ------------------------------------------------------------
// 与玩家相关的网络连接
var (
	g_clients     = make(map[uint32]*tcp.TCPConn) //accountId-clientConn
	g_client_game = make(map[uint32]int)          //accountId-gameSvrId
	g_game_online = make(map[int]int32)           //gameSvrId-online
)

func AddClientConn(aid uint32, conn *tcp.TCPConn) { g_clients[aid] = conn }
func DelClientConn(aid uint32)                    { delete(g_clients, aid) }
func GetClientConn(aid uint32) *tcp.TCPConn       { return g_clients[aid] }

func AddGameConn(aid uint32, svrId int) { g_client_game[aid] = svrId; g_game_online[svrId]++ }
func DelGameConn(aid uint32) {
	if svrId, ok := g_client_game[aid]; ok {
		if n, ok := g_game_online[svrId]; ok {
			g_game_online[svrId] = n - 1
		}
	}
	delete(g_client_game, aid)
}
func GetGameConn(aid uint32) *tcp.TCPConn { return netConfig.GetGameConn(g_client_game[aid]) }
