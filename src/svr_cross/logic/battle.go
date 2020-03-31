package logic

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/tcp"
)

const (
	K_Player_Max  = 5000 //战斗服玩家容量
	K_Player_Base = 2000 //人数填够之后，各服均分
)

var (
	g_battle_player_cnt = make(map[int]uint32) //各服在线人数
	g_cur_select_idx    = -1
)

// ------------------------------------------------------------
// - 将svr_game中的玩家属性数据发到svr_battle
func Rpc_cross_relay_battle_data(req, _ *common.NetPack, conn *tcp.TCPConn) {
	version := req.ReadString()
	args := req.LeftBuf()
	svrId := _SelectBattleSvrId(version)
	if svrId == -1 {
		//TODO:zhoumf: 通报client战斗服已满
		gamelog.Info("svr_battle is full !!!")
		return
	}
	gamelog.Debug("select battle: %d", svrId)

	oldReqKey := req.GetReqKey()
	netConfig.CallRpcBattle(svrId, enum.Rpc_battle_handle_player_data, func(buf *common.NetPack) {
		buf.WriteBuf(args)
	}, func(backBuf *common.NetPack) {
		playerCnt := backBuf.ReadUInt32() //选中战斗服的已有人数
		g_battle_player_cnt[svrId] = playerCnt
		//异步回调，不能直接用ack
		backBuf.SetReqKey(oldReqKey)
		conn.WriteMsg(backBuf)
	})
}
func _SelectBattleSvrId(version string) int {
	//moba类的，应该有个专门的匹配服，供自由玩家【快速】组房间
	//io向的，允许中途加入，应尽量分配到人多的战斗服
	vs := meta.GetMetas("battle", version)
	//1、优先在各个服务器分配一定人数
	for _, v := range vs {
		if g_battle_player_cnt[v.SvrID] < K_Player_Base {
			return v.SvrID
		}
	}
	//2、基础人数够了，再各服均分
	for i := 0; i < len(vs); i++ {
		//先自增判断越界，防止中途有战斗服宕机
		if g_cur_select_idx++; g_cur_select_idx >= len(vs) {
			g_cur_select_idx = 0
		}
		v := vs[g_cur_select_idx]
		if g_battle_player_cnt[v.SvrID] < K_Player_Max {
			return v.SvrID
		}
	}
	return -1
}

// ------------------------------------------------------------
// - 转至svr_game
func Rpc_cross_relay_to_game(req, ack *common.NetPack, conn *tcp.TCPConn) {
	svrId := req.ReadInt()
	args := req.LeftBuf() //头两字段须是：rid、aid
	if p, ok := netConfig.GetGameRpc(svrId); ok {
		p.CallRpc(enum.Rpc_recv_player_msg, func(buf *common.NetPack) {
			buf.WriteBuf(args)
		}, nil)
	}
}
