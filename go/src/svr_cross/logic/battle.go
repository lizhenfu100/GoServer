package logic

import (
	"common"
	"svr_cross/api"
	"tcp"
)

//////////////////////////////////////////////////////////////////////
//!
func Rpc_Relay_Battle_Data(conn *tcp.TCPConn, msg *common.NetPack) {

	// 转发给Battle进程
	msg.SetRpc("rpc_handle_battle_data")
	api.SendToBattle(1, msg)

	//设置定时器，
	//若还未收到匹配成功消息，30s后通知战斗服取消
	//须等战斗服的取消回应，取消成功，则将本批次人员挪至下个战斗服继续
}
