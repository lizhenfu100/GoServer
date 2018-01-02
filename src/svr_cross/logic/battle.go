package logic

import (
	"common"
	"common/net/meta"
	"gamelog"
	"generate_out/rpc/enum"
	"sort"
	"svr_cross/api"
	"tcp"
)

const (
	K_Player_Limit = 5000 //战斗服玩家容量
	K_Player_Base  = 2000 //人数填够之后，各服均分
)

var (
	g_battle_player_cnt = make(map[int]uint32) //各服在线人数
	g_cur_select_idx    = -1
)

// ------------------------------------------------------------
//! TODO：某些 rpc 需告知 Battle 来源的 GameSvrID
func Rpc_cross_relay_battle_data(req, ack *common.NetPack, conn *tcp.TCPConn) {
	svrId := _SelectBattleSvrId()
	if svrId == -1 {
		//FIXME:无空闲战斗服时，自动执行脚本，开新战斗服(怎么开?)
		gamelog.Error("!!! svr_battle is full !!!")
		return
	}
	gamelog.Debug("select battle: %d", svrId)

	// 转给Battle进程
	api.CallRpcBattle(svrId, enum.Rpc_battle_handle_player_data, func(buf *common.NetPack) {
		buf.WriteBuf(req.Body())
	}, func(backBuf *common.NetPack) {
		playerCnt := backBuf.ReadUInt32() //选中战斗服的已有人数
		g_battle_player_cnt[svrId] = playerCnt

		//【Notice：异步回调里不能用非线程安全的数据，直接用ack回复错的】
		gamelog.Debug("send addr to game")
		ip, port := meta.GetIpPort("battle", svrId)
		gameMsg := common.NewNetPackCap(256)
		gameMsg.SetOpCode(enum.Rpc_game_battle_ack)
		gameMsg.WriteString(ip)
		gameMsg.WriteUInt16(port)
		gameMsg.WriteBuf(backBuf.LeftBuf()) //[]<pid>
		conn.WriteMsg(gameMsg)
		gameMsg.Free()
	})
}
func _SelectBattleSvrId() int {
	//moba类的，应该有个专门的匹配服，供自由玩家【快速】组房间
	//io向的，允许中途加入，应尽量分配到人多的战斗服
	ids := meta.GetModuleIDs("battle")
	sort.Ints(ids)
	//1、优先在各个服务器分配一定人数
	for i := 0; i < len(ids); i++ {
		if g_battle_player_cnt[ids[i]] < K_Player_Base {
			return ids[i]
		}
	}
	//2、基础人数够了，再各服均分
	for i := 0; i < len(ids); i++ {
		//先自增判断越界，防止中途有战斗服宕机
		if g_cur_select_idx++; g_cur_select_idx >= len(ids) {
			g_cur_select_idx = 0
		}
		svrId := ids[g_cur_select_idx]
		if g_battle_player_cnt[svrId] < K_Player_Limit {
			return svrId
		}
	}
	return -1
}
func Rpc_cross_echo(req, ack *common.NetPack, conn *tcp.TCPConn) {
	str := req.ReadString()
	print(str)
	ack.WriteString(str)
}
