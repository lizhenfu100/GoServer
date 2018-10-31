/***********************************************************************
* @ game svr list
* @ brief
	玩家手选区服：
		1、下发区服列表，client选定具体gamesvr节点登录
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
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/gameInfo"
	"sync"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) {
	version := req.ReadString()
	//临时读取buffer数据
	oldPos := req.ReadPos
	gameName := req.ReadString()
	//gameSvrId := req.ReadInt() //若手选区服方式，则登录时自己上报节点编号
	account := req.ReadString()
	req.ReadPos = oldPos

	//同步验证账号信息，取得绑定的gameInfo
	centerSvrId := netConfig.HashCenterID(account)
	netConfig.SyncRelayToCenter(centerSvrId, enum.Rpc_center_account_login, req, ack)

	//临时读取buffer数据
	oldPos = ack.ReadPos
	errCode := ack.ReadUInt16()
	if errCode != err.Success {
		ack.ReadPos = oldPos
	} else {
		accountId := ack.ReadUInt32()
		//println("---------------- Rpc_login_account_login: ", accountId)
		var gameInfo2 gameInfo.TGameInfo
		gameInfo2.BufToData(ack)
		ack.Clear()

		gameSvrId := gameInfo2.SvrId
		if gameSvrId == 0 { //自动选取空闲节点，并绑定到账号上
			if gameSvrId = GetFreeGameSvrId(version); gameSvrId > 0 {
				netConfig.CallRpcCenter(centerSvrId, enum.Rpc_center_set_game_info, func(buf *common.NetPack) {
					buf.WriteUInt32(accountId)
					buf.WriteString(gameName)
					gameInfo2.SvrId = gameSvrId
					gameInfo2.DataToBuf(buf)
				}, nil)
			}
		}
		if gameSvrId <= 0 {
			ack.WriteUInt16(err.None_free_game_server)
			return
		}

		//登录成功，设置各类信息
		gatewayId := netConfig.HashGatewayID(accountId) //此玩家要连接的gateway

		//生成一个临时token，发给gamesvr、client用以登录验证
		token := atomic.AddUint32(&g_login_token, 1)
		var pMetaToClient *meta.Meta //回复client要连接的目标节点
		//【Notice: CallRpc接口不是线程安全的，http后台不适用】
		if conn := netConfig.GetTcpConn("gateway", gatewayId); conn != nil {
			msg := common.NewNetPackCap(32)
			msg.SetOpCode(enum.Rpc_gateway_login_token)
			msg.WriteUInt32(token)
			msg.WriteUInt32(accountId)
			msg.WriteInt(gameSvrId)
			conn.WriteMsg(msg)

			pMetaToClient = meta.GetMeta("gateway", gatewayId)
			errCode = err.None_gateway

		} else if conn := netConfig.GetTcpConn("game", gameSvrId); conn != nil {
			msg := common.NewNetPackCap(32)
			msg.SetOpCode(enum.Rpc_game_login_token)
			msg.WriteUInt32(token)
			msg.WriteUInt32(accountId)
			conn.WriteMsg(msg)

			pMetaToClient = meta.GetMeta("game", gameSvrId)
			errCode = err.None_game_server

		} else if addr := netConfig.GetHttpAddr("game", gameSvrId); addr != "" {
			http.CallRpc(addr, enum.Rpc_game_login_token, func(buf *common.NetPack) {
				buf.WriteUInt32(token)
				buf.WriteUInt32(accountId)
			}, nil)

			pMetaToClient = meta.GetMeta("game", gameSvrId)
			errCode = err.None_game_server
		}

		if pMetaToClient != nil { //回复client要连接的目标节点
			ack.WriteUInt16(err.Success)
			ack.WriteUInt32(accountId)
			ack.WriteString(pMetaToClient.OutIP)
			ack.WriteUInt16(pMetaToClient.Port())
			ack.WriteUInt32(token)
		} else {
			ack.WriteUInt16(errCode)
		}
	}
}
func Rpc_login_relay_to_center(req, ack *common.NetPack) {
	rpcId := req.ReadUInt16() //目标rpc
	//临时读取buffer数据
	oldPos := req.ReadPos
	strKey := req.ReadString() //accountName/bindVal
	req.ReadPos = oldPos

	svrId := netConfig.HashCenterID(strKey)
	netConfig.SyncRelayToCenter(svrId, rpcId, req, ack)
}

// -------------------------------------
// game svr list
var g_game_player_cnt sync.Map //gameSvrId-playerCnt

func Rpc_login_get_gamesvr_list(req, ack *common.NetPack) {
	version := req.ReadString()

	ids, _ := meta.GetModuleIDs("game", version)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		pMeta := meta.GetMeta("game", id)
		ack.WriteInt(id)
		ack.WriteString(pMeta.SvrName)
		ack.WriteString(pMeta.OutIP)
		ack.WriteUInt16(pMeta.Port())
		playerCnt := int32(0)
		if v, ok := g_game_player_cnt.Load(id); ok {
			playerCnt = v.(int32)
		}
		ack.WriteInt32(playerCnt)
	}
}
func Rpc_login_set_player_cnt(req, ack *common.NetPack) {
	svrId := req.ReadInt()
	cnt := req.ReadInt32()
	g_game_player_cnt.Store(svrId, cnt)
}
func GetFreeGameSvrId(version string) int {
	ret, minCnt := -1, int32(999999)
	ids, _ := meta.GetModuleIDs("game", version)
	for _, id := range ids {
		if pMeta := meta.GetMeta("game", id); pMeta != nil && pMeta.IsMatchVersion(version) {
			if v, ok := g_game_player_cnt.Load(id); ok {
				if cnt := v.(int32); cnt < minCnt {
					ret, minCnt = id, cnt
				}
			} else {
				ret, minCnt = id, 0
			}
		}
	}
	return ret
}
