package logic

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
)

func Rpc_gateway_relay_player_msg(req *common.NetPack, recvFun func(*common.NetPack)) {
	rpcId := req.ReadUInt16()     //目标rpc
	accountId := req.ReadUInt32() //目标玩家
	gamelog.Debug("relay_player_msg(%d)", rpcId)

	if accountId == 0 {
		gamelog.Error("accountId nil")
	} else if netConfig.HashGatewayID(accountId) == meta.G_Local.SvrID { //应连本节点的玩家
		RelayPlayerMsg(rpcId, accountId, req.LeftBuf(), recvFun)
	} else { //非本节点玩家，转至其它gateway
		netConfig.CallRpcGateway(accountId, rpcId, func(buf *common.NetPack) {
			buf.WriteBuf(req.LeftBuf())
		}, recvFun)
	}
}
func RelayPlayerMsg(rpcId uint16, accountId uint32, reqData []byte, recvFun func(*common.NetPack)) {
	rpcModule := enum.GetRpcModule(rpcId)
	if rpcModule == "" {
		gamelog.Error("rpc(%d) havn't module", rpcId)
		return
	}
	if rpcModule == "client" { //转给客户端
		if p := GetClientConn(accountId); p != nil {
			p.CallRpcSafe(rpcId, func(buf *common.NetPack) {
				buf.WriteBuf(reqData)
			}, recvFun)
		} else {
			gamelog.Error("rid(%d) accountId(%d) client conn nil", rpcId, accountId)
		}
	} else { //转给后台节点
		if p, ok := GetModuleRpc(rpcModule, accountId); ok {
			p.CallRpcSafe(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
				buf.WriteUInt16(rpcId)
				buf.WriteUInt32(accountId)
				buf.WriteBuf(reqData)
			}, recvFun)
		} else {
			//【系统缺陷】丢失了玩家的game路由，没通知调用者，也没重试
			gamelog.Error("rid(%d) accountId(%d) svr conn nil", rpcId, accountId)
		}
	}
}
