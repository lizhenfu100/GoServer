package logic

import (
	"common"
	"gamelog"
	"netConfig"
	"svr_cross/api"
	"tcp"
)

const (
	K_Player_Cap = 5000 //战斗服玩家容量
)

var (
	g_battle_player_cnt = make(map[int]uint32)
)

//////////////////////////////////////////////////////////////////////
//!
func Rpc_Relay_Battle_Data(req, ack *common.NetPack, conn *tcp.TCPConn) {
	//moba类的，应该有个专门的匹配服，供自由玩家【快速】组房间
	//io向的，允许中途加入，应尽量分配到人多的战斗服
	svrId := 0
	ids := tcp.GetRegModuleIDs("battle")
	for i := 0; i < len(ids); i++ {
		if g_battle_player_cnt[ids[i]] < K_Player_Cap {
			svrId = ids[i]
			break
		}
	}
	if svrId == 0 {
		//FIXME:无空闲战斗服时，自动执行脚本，开新战斗服(怎么开?)
		gamelog.Error("!!! svr_battle is full !!!")
		return
	}
	// 转给Battle进程
	api.GetBattleConn(svrId).CallRpcSafe("rpc_battle_handle_player_data", func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(backBuf *common.NetPack) {
		// playerCnt := backBuf.ReadUInt32() //选中战斗服的已有人数
		// g_battle_player_cnt[svrId] = playerCnt

		//【Notice：异步回调里不能用非线程安全的数据，直接用ack回复错的】
		print("--- send addr to game ---\n")
		gameMsg := common.NewNetPackCap(256)
		gameMsg.SetRpc("rpc_game_battle_ack")
		netConfig.WriteAddr(gameMsg, "battle", &svrId) //string uint16
		gameMsg.WriteBuf(backBuf.Body())               //[]<pid>
		conn.WriteMsg(gameMsg)
	})
}
