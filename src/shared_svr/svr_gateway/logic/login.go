/***********************************************************************
* @ game svr list
* @ brief
	玩家手选区服：
		1、gateway下发区服列表，client选定具体gamesvr节点登录
		2、client本地存储gamesvr节点编号，登录时上报

	后台自动分服：
		1、验证账号，若账号中无绑定区服信息，进入创建角色流程
		2、选择最空闲gamesvr节点，创建角色成功，将区服编号绑账号上
		3、下次登录，验证账号后即定位具体gamesvr节点

* @ author zhoumf
* @ date 2018-3-19
***********************************************************************/
package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/gameInfo/SoulKnight"
	"tcp"
)

// ------------------------------------------------------------
// 注册、登录
//TODO:zhoumf: 与svr_center的交互是同步的，放gateway会影响整体消息处理性能(tcp_rpc单线程的)
//TODO:zhoumf: Rpc_gateway_account_login  Rpc_gateway_relay_to_center
func Rpc_gateway_relay_to_center(req, ack *common.NetPack, client *tcp.TCPConn) {
	rpcId := req.ReadUInt16() //目标rpc
	//临时读取buffer数据
	oldPos := req.ReadPos
	strKey := req.ReadString() //accountName/bindVal
	req.ReadPos = oldPos

	svrId := netConfig.HashCenterID(strKey)
	netConfig.SyncRelayToCenter(svrId, rpcId, req, ack)
}
func Rpc_gateway_account_login(req, ack *common.NetPack, client *tcp.TCPConn) { //req: 客户端版本号，游戏名，账号，密码
	version := req.ReadString()
	//临时读取buffer数据
	oldPos := req.ReadPos
	gameName := req.ReadString()
	account := req.ReadString()
	//gameSvrId := req.ReadInt() //若手选区服方式，则登录时自己上报节点编号
	req.ReadPos = oldPos

	//验证账号信息，取得绑定的gameInfo
	centerSvrId := netConfig.HashCenterID(account)
	netConfig.SyncRelayToCenter(centerSvrId, enum.Rpc_center_account_login, req, ack)

	//临时读取buffer数据
	oldPos = ack.ReadPos
	defer func() { ack.ReadPos = oldPos }()
	if errCode := ack.ReadInt8(); errCode > 0 {
		accountId := ack.ReadUInt32()
		var gameInfo SoulKnight.TGameInfo //解析账号上的gameInfo
		gameInfo.BufToData(ack)

		//自动选取空闲节点，并绑定到账号上
		if gameInfo.SvrId == 0 {
			if gameInfo.SvrId = GetFreeSvrId(version); gameInfo.SvrId > 0 {
				netConfig.CallRpcCenter(centerSvrId, enum.Rpc_center_set_game_info, func(buf *common.NetPack) {
					buf.WriteUInt32(accountId)
					buf.WriteString(gameName)
					gameInfo.DataToBuf(buf)
				}, nil)
			}
		}

		//登录成功，查看目标节点在线人数
		gameSvrId := gameInfo.SvrId
		if UpdateGamePlayerCnt(gameSvrId) {
			//设置此玩家的game路由
			client.UserPtr = accountId
			AddClientConn(accountId, client)
			AddGameConn(accountId, gameSvrId)
		}
	}
}
func Rpc_gateway_game_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if pConn := GetGameConn(accountId); pConn != nil {
			oldReqKey := req.GetReqKey()
			pConn.CallRpc(enum.Rpc_game_login, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}
func Rpc_gateway_game_create_player(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if pConn := GetGameConn(accountId); pConn != nil {
			oldReqKey := req.GetReqKey()
			pConn.CallRpc(enum.Rpc_game_create_player, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}

// ------------------------------------------------------------
// game svr list
var g_game_player_cnt = make(map[int]int32) //gameSvrId-playerCnt

func Rpc_gateway_get_gamesvr_list(req, ack *common.NetPack, client *tcp.TCPConn) {
	version := req.ReadString()

	ids, _ := meta.GetModuleIDs("game", version)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		pMeta := meta.GetMeta("game", id)
		ack.WriteInt(id)
		ack.WriteString(pMeta.SvrName)
		ack.WriteString(pMeta.OutIP)
		ack.WriteUInt16(pMeta.Port())
		ack.WriteInt32(g_game_player_cnt[id])
	}
}
func GetFreeSvrId(version string) int {
	if len(g_game_player_cnt) == 0 {
		ids, _ := meta.GetModuleIDs("game", version)
		for _, id := range ids {
			g_game_player_cnt[id] = 0
		}
	}
	var pMeta *meta.Meta
	var minCnt = int32(999999)
	for id, cnt := range g_game_player_cnt {
		if cnt < minCnt {
			if pMeta = meta.GetMeta("game", id); pMeta != nil && pMeta.IsMatchVersion(version) {
				minCnt = cnt
			} else {
				pMeta = nil
			}
		}
	}
	if pMeta != nil {
		return pMeta.SvrID
	} else {
		return -1
	}
}
func UpdateGamePlayerCnt(svrId int) bool {
	if pConn := netConfig.GetGameConn(svrId); pConn != nil {
		pConn.CallRpc(enum.Rpc_game_get_player_cnt, func(buf *common.NetPack) {

		}, func(backBuf *common.NetPack) {
			cnt := backBuf.ReadInt32()
			g_game_player_cnt[svrId] = cnt
		})
		return true
	}
	return false
}
