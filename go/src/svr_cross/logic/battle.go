package logic

import (
	"common"
	"netConfig"
	"svr_cross/api"
	"tcp"
)

//////////////////////////////////////////////////////////////////////
//!
func Rpc_Relay_Battle_Data(req, ack *common.NetPack, conn *tcp.TCPConn) {
	//TODO:zhoumf: 挑选战斗服--战斗服及时反馈在线人数，达到上限的跳过
	//无空闲战斗服时，自动执行脚本，开新战斗服(怎么开?)

	// 转给Battle进程
	svrId := 1
	api.GetBattleConn(svrId).CallRpcSafe("rpc_handle_battle_data", func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(backBuf *common.NetPack) {
		//【Notice：异步回调里不能用非线程安全的数据，直接用ack回复错的】
		print("--- send addr to game ---\n")
		gameMsg := common.NewNetPackCap(256)
		gameMsg.SetRpc("rpc_battle_ack")
		netConfig.WriteAddr(gameMsg, "battle", &svrId) //string uint16
		gameMsg.WriteBuf(backBuf.Body())               //[]<pid, idx>
		conn.WriteMsg(gameMsg)
	})
}
