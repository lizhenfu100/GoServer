package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"svr_login/api"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) {
	_AckLoginAndCreateToken(req, ack, enum.Rpc_center_account_login)
}
func Rpc_center_bind_info_login(req, ack *common.NetPack) {
	_AckLoginAndCreateToken(req, ack, enum.Rpc_center_bind_info_login)
}
func _AckLoginAndCreateToken(req, ack *common.NetPack, rid uint16) {
	svrId := req.ReadInt()

	//Notice: 这里必须是同步调用，CallRpc的回调里才能有效使用ack
	//由于svr_center、svr_login都是http服务器，所以能这么搞
	//如果是tcp服务，就得分成两条消息，让login主动通知client。参考svr_cross->battle.go中的做法，CallRpcCenter的recvBuf里，操作TcpConn通知，不通过ack
	isSyncCall := false
	api.CallRpcCenter(rid, func(buf *common.NetPack) {
		buf.WriteBuf(req.LeftBuf())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		err := recvBuf.ReadInt8()
		if err > 0 {
			accountId := recvBuf.ReadUInt32()
			if ip, port := netConfig.GetIpPort("game", svrId); port <= 0 {
				ack.WriteInt8(-100) //invalid_svrid
			} else {
				ack.WriteInt8(1)
				ack.WriteString(ip)
				ack.WriteUInt16(port)
				ack.WriteUInt32(accountId)
				//生成一个临时token，发给gamesvr、client，用以登录验证
				token := atomic.AddUint32(&g_login_token, 1)
				ack.WriteUInt32(token)

				api.CallRpcGame(svrId, enum.Rpc_game_login_token, func(buf *common.NetPack) {
					buf.WriteUInt32(accountId)
					buf.WriteUInt32(token)
				}, nil)
			}
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
	api.SyncRelayToCenter(enum.Rpc_center_reg_account, req, ack)
}
func Rpc_login_check_account(req, ack *common.NetPack) {
	api.SyncRelayToCenter(enum.Rpc_center_check_account, req, ack)
}
func Rpc_login_change_password(req, ack *common.NetPack) {
	api.SyncRelayToCenter(enum.Rpc_center_change_password, req, ack)
}
func Rpc_login_bind_info(req, ack *common.NetPack) {
	api.SyncRelayToCenter(enum.Rpc_center_bind_info, req, ack)
}
func Rpc_login_get_account_by_bind_info(req, ack *common.NetPack) {
	api.SyncRelayToCenter(enum.Rpc_center_get_account_by_bind_info, req, ack)
}
