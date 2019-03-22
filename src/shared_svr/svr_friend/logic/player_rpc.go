package logic

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"nets/tcp"
)

type PlayerRpc func(req, ack *common.NetPack, this *TFriendModule)

var G_PlayerHandleFunc [enum.RpcEnumCnt]PlayerRpc

func RegPlayerRpc(list map[uint16]PlayerRpc) {
	for k, v := range list {
		G_PlayerHandleFunc[k] = v
	}
}
func DoPlayerRpc(this *TFriendModule, rpcId uint16, req, ack *common.NetPack) bool {
	if msgFunc := G_PlayerHandleFunc[rpcId]; msgFunc != nil {
		msgFunc(req, ack, this)
		return true
	}
	gamelog.Error("Msg(%d) Not Regist", rpcId)
	return false
}

// ------------------------------------------------------------
// - 网关转发的玩家消息
func Rpc_recv_player_msg(req, ack *common.NetPack, conn *tcp.TCPConn) {
	rpcId := req.ReadUInt16()
	accountId := req.ReadUInt32()

	gamelog.Debug("PlayerMsg:%d", rpcId)

	if this := FindWithDB(accountId); this != nil {
		DoPlayerRpc(this, rpcId, req, ack)
	} else {
		gamelog.Debug("Player(%d) isn't online", accountId)
	}
}
