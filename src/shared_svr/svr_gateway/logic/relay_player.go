package logic

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/tcp"
	"sync"
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
		if p := GetClientConn(accountId); p != nil {
			p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(rpcData)
			}, func(backBuf *common.NetPack) {
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		} else {
			gamelog.Error("rid(%d) accountId(%d) client conn nil", rpcId, accountId)
		}
	} else { //转给后台节点
		if p, ok := _GetModuleRpc(rpcModule, accountId); ok {
			p.CallRpcSafe(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
				buf.WriteUInt16(rpcId)
				buf.WriteUInt32(accountId)
				buf.WriteBuf(rpcData)
			}, func(backBuf *common.NetPack) {
				backBuf.SetReqKey(oldReqKey)
				conn.WriteMsg(backBuf)
			})
		} else {
			gamelog.Error("rid(%d) accountId(%d) svr conn nil", rpcId, accountId)
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
func _GetModuleRpc(module string, aid uint32) (netConfig.Rpc, bool) {
	if module == "game" {
		return GetGameRpc(aid)
	} else {
		// 其它节点无状态的，AccountId取模得节点id
		ids := meta.GetModuleIDs(module, meta.G_Local.Version)
		if length := uint32(len(ids)); length > 0 {
			id := ids[aid%length]
			return netConfig.GetRpc(module, id)
		}
	}
	return nil, false
}

// ------------------------------------------------------------
// 与玩家相关的网络连接
var (
	g_clients = make(map[uint32]*tcp.TCPConn) //accountId-clientConn
	//TODO:zhoumf:gateway重启，得重登录收集该信息，挫 …… 把<aid, gameSvr>扔redis里，抽离出gateway
	//可以研究一下orleans，灵感的源泉
	//状态抽离到单独节点，统一记录
	//感觉轮询的方式也挺不错，gateway收到消息，问询下后面的节点可否处理，并缓存结果信息，只玩家上来时会轮询次，后面就是固定路由了
	g_client_game sync.Map //accountId-gameSvrId
)

func AddClientConn(aid uint32, conn *tcp.TCPConn) { g_clients[aid] = conn }
func DelClientConn(aid uint32)                    { delete(g_clients, aid) }
func GetClientConn(aid uint32) *tcp.TCPConn       { return g_clients[aid] }

func AddPlayer(aid uint32, svrId int) { g_client_game.Store(aid, svrId) }
func DelPlayer(aid uint32)            { g_client_game.Delete(aid) }

func GetGameRpc(aid uint32) (netConfig.Rpc, bool) {
	if v, ok := g_client_game.Load(aid); ok {
		return netConfig.GetGameRpc(v.(int))
	}
	return nil, false
}
