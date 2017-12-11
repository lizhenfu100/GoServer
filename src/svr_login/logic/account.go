package logic

import (
	"common"
	"common/net/meta"
	"generate_out/rpc/enum"
	"svr_login/api"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) {
	_AckLoginAndCreateToken(req, ack, enum.Rpc_center_account_login)
}
func Rpc_login_bind_info_login(req, ack *common.NetPack) {
	_AckLoginAndCreateToken(req, ack, enum.Rpc_center_bind_info_login)
}
func _AckLoginAndCreateToken(req, ack *common.NetPack, rid uint16) {
	key := req.ReadString()
	gameSvrId := req.ReadInt()
	centerSvrId := api.HashCenterID(key)

	ip, port := meta.GetIpPort("game", gameSvrId)
	if port <= 0 {
		ack.WriteInt8(-100) //invalid_svrid
		return
	}
	//Notice: 这里必须是同步调用，CallRpc的回调里才能有效使用ack
	//由于svr_center、svr_login都是http服务器，所以能这么搞
	//如果是tcp服务，就得分成两条消息，让login主动通知client。参考svr_cross->battle.go中的做法，CallRpcCenter的recvBuf里，操作TcpConn通知，不通过ack
	isSyncCall := false
	api.CallRpcCenter(centerSvrId, rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.LeftBuf())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		err := recvBuf.ReadInt8()
		if err > 0 {
			//生成一个临时token，发给gamesvr、client，用以登录验证
			token := atomic.AddUint32(&g_login_token, 1)
			accountId := recvBuf.ReadUInt32()
			ack.WriteInt8(1)
			ack.WriteString(ip)
			ack.WriteUInt16(port)
			ack.WriteUInt32(accountId)
			ack.WriteUInt32(token)
			//将token发给目标gamesvr
			api.CallRpcGame(gameSvrId, enum.Rpc_game_login_token, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteUInt32(token)
			}, nil)
		} else {
			ack.WriteInt8(err)
		}
	})
	if isSyncCall == false {
		panic("Using ack int another CallRpc must be sync!!! zhoumf\n")
	}
}

// -------------------------------------
// svr_center agent
func Rpc_login_reg_account(req, ack *common.NetPack) {
	account := req.ReadString()
	req.ReadPos -= (2 + len([]byte(account)))
	svrId := api.HashCenterID(account)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_reg_account, req, ack)
}
func Rpc_login_check_account(req, ack *common.NetPack) {
	account := req.ReadString()
	req.ReadPos -= (2 + len([]byte(account)))
	svrId := api.HashCenterID(account)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_check_account, req, ack)
}
func Rpc_login_change_password(req, ack *common.NetPack) {
	account := req.ReadString()
	req.ReadPos -= (2 + len([]byte(account)))
	svrId := api.HashCenterID(account)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_change_password, req, ack)
}
func Rpc_login_bind_info(req, ack *common.NetPack) {
	account := req.ReadString()
	req.ReadPos -= (2 + len([]byte(account)))
	svrId := api.HashCenterID(account)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_bind_info, req, ack)
}
func Rpc_login_get_account_by_bind_info(req, ack *common.NetPack) {
	bindKey := req.ReadString()
	bindVal := req.ReadString()
	req.ReadPos -= (4 + len([]byte(bindKey)) + len([]byte(bindVal)))
	svrId := api.HashCenterID(bindVal)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_get_account_by_bind_info, req, ack)
}
func Rpc_login_get_accountid(req, ack *common.NetPack) {
	account := req.ReadString()
	req.ReadPos -= (2 + len([]byte(account)))
	svrId := api.HashCenterID(account)
	api.SyncRelayToCenter(svrId, enum.Rpc_center_account_login, req, ack)
}