package logic

import (
	"common"
	"gamelog"
	"netConfig"
	"sort"
	"svr_cross/api"
	"tcp"
)

const (
	K_Player_Cap   = 5000 //战斗服玩家容量
	K_Player_Limit = 2000 //人数填够之后，各服均分
)

var (
	g_battle_player_cnt = make(map[int]uint32)
	g_cur_select_idx    = 0
)

//////////////////////////////////////////////////////////////////////
//!
func Rpc_Relay_Battle_Data(req, ack *common.NetPack, conn *tcp.TCPConn) {
	svrId := _SelectBattleSvrId()
	if svrId == 0 {
		//FIXME:无空闲战斗服时，自动执行脚本，开新战斗服(怎么开?)
		gamelog.Error("!!! svr_battle is full !!!")
		return
	}
	// 转给Battle进程
	api.CallRpcBattle(svrId, "rpc_battle_handle_player_data", func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(backBuf *common.NetPack) {
		playerCnt := backBuf.ReadUInt32() //选中战斗服的已有人数
		g_battle_player_cnt[svrId] = playerCnt

		//【Notice：异步回调里不能用非线程安全的数据，直接用ack回复错的】
		print("--- send addr to game ---\n")
		ip, port := netConfig.GetIpPort("battle", svrId)
		gameMsg := common.NewNetPackCap(256)
		gameMsg.SetRpc("rpc_game_battle_ack")
		gameMsg.WriteString(ip)
		gameMsg.WriteUInt16(port)
		gameMsg.WriteBuf(backBuf.LeftBuf()) //[]<pid>
		conn.WriteMsg(gameMsg)
	})
}
func _SelectBattleSvrId() int {
	//moba类的，应该有个专门的匹配服，供自由玩家【快速】组房间
	//io向的，允许中途加入，应尽量分配到人多的战斗服
	ids := tcp.GetRegModuleIDs("battle")
	sort.Ints(ids)
	//1、优先在各个服务器分配一定人数
	for i := 0; i < len(ids); i++ {
		if g_battle_player_cnt[ids[i]] < K_Player_Limit {
			return ids[i]
		}
	}
	//2、基础人数够了，再各服均分
	isFull := true
	for i := 0; i < len(ids); i++ {
		if g_battle_player_cnt[ids[i]] < K_Player_Cap {
			isFull = false
		}
	}
	if isFull == false {
		for i := 0; i < len(ids); i++ {
			svrId := ids[g_cur_select_idx]

			g_cur_select_idx++
			if g_cur_select_idx >= len(ids) {
				g_cur_select_idx = 0
			}

			if g_battle_player_cnt[svrId] < K_Player_Cap {
				return svrId
			}
		}
	}
	return 0
}
