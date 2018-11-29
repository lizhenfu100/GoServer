/***********************************************************************
* @ game svr list
* @ brief
	登录服决议：
		1、客户端显示大区列表，玩家自己选择
		2、login iplist 写成前端配置，玩家注册时挨个ping，取延时最小的去注册

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
	//1、临时读取buffer数据，排除版本号，将数据转发center
	oldPos := req.ReadPos
	gameName := req.ReadString()
	//gameSvrId := req.ReadInt() //若手选区服则自己上报gameSvrId
	account := req.ReadString()
	req.ReadPos = oldPos

	centerSvrId := netConfig.HashCenterID(account)

	//2、同步转至center验证账号信息，取得accountId、gameInfo
	if ok, accountId, gameInfo2 := accountLogin2(centerSvrId, req, ack); ok {
		//3、选取gameSvrId；若手选区服则由客户端上报
		gameSvrId := accountLogin3(accountId, &gameInfo2, version, gameName, centerSvrId)
		if gameSvrId <= 0 {
			ack.WriteUInt16(err.None_free_game_server)
		} else {
			//4、登录成功，设置各类信息，回复client待连接的节点(gateway或game)
			accountLogin4(accountId, gameSvrId, ack)
		}
	}
}
func accountLogin2(centerSvrId int, req, ack *common.NetPack) (_ok bool, _aid uint32, _info gameInfo.TGameInfo) {
	//同步至center验证账号信息，取得accountId、gameInfo
	netConfig.SyncRelayToCenter(centerSvrId, enum.Rpc_center_account_login, req, ack)

	//临时读取buffer数据
	oldPos := ack.ReadPos
	if ack.ReadUInt16() != err.Success {
		ack.ReadPos = oldPos
		_ok = false
	} else {
		_aid = ack.ReadUInt32()
		_info.BufToData(ack)
		ack.Clear()
		_ok = true
	}
	return
}
func accountLogin3(accountId uint32, pInfo *gameInfo.TGameInfo, version, gameName string, centerSvrId int) (_gameSvrId int) {
	//选取gameSvrId：若账户未绑定游戏服，自动选取空闲节点，并绑定到账号上
	if _gameSvrId = pInfo.SvrId; _gameSvrId == 0 {
		if _gameSvrId = GetFreeGameSvrId(version); _gameSvrId > 0 {
			netConfig.CallRpcCenter(centerSvrId, enum.Rpc_center_set_game_info, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteString(gameName)
				pInfo.SvrId = _gameSvrId
				pInfo.DataToBuf(buf)
			}, nil)
		}
	}
	return
}
func accountLogin4(accountId uint32, gameSvrId int, ack *common.NetPack) {
	//登录成功，设置各类信息，回复client待连接的节点(gateway或game)
	gatewayId := netConfig.HashGatewayID(accountId) //此玩家要连接的gateway
	errCode := err.Unknow_error

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
		msg.Free()

		pMetaToClient = meta.GetMeta("gateway", gatewayId)
		errCode = err.None_gateway

	} else if conn := netConfig.GetTcpConn("game", gameSvrId); conn != nil {
		msg := common.NewNetPackCap(32)
		msg.SetOpCode(enum.Rpc_game_login_token)
		msg.WriteUInt32(token)
		msg.WriteUInt32(accountId)
		conn.WriteMsg(msg)
		msg.Free()

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

// ------------------------------------------------------------
// 转发至center
func Rpc_login_relay_to_center(req, ack *common.NetPack) {
	rpcId := req.ReadUInt16() //目标rpc
	//临时读取buffer数据
	oldPos := req.ReadPos
	strKey := req.ReadString() //accountName/bindVal
	req.ReadPos = oldPos

	svrId := netConfig.HashCenterID(strKey)
	netConfig.SyncRelayToCenter(svrId, rpcId, req, ack)
}

// ------------------------------------------------------------
// 玩家从登录服取其它节点信息
func Rpc_login_get_meta_list(req, ack *common.NetPack) {
	module := req.ReadString() //game、save、file...
	version := req.ReadString()

	ids, _ := meta.GetModuleIDs(module, version)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		p := meta.GetMeta(module, id)
		ack.WriteInt(id)
		ack.WriteString(p.SvrName)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())

		switch module { //模块节点的附带信息
		case "game":
			if v, ok := g_game_player_cnt.Load(id); ok {
				ack.WriteInt32(v.(int32))
			} else {
				ack.WriteInt32(0)
			}
		}
	}
}

// 在线人数：精确性要求不高，让各节点主动周期性上报
//【维护不同进程的数据一致，须保证“增删改查”，成本过高】
var g_game_player_cnt sync.Map //<gameSvrId, playerCnt>

func Rpc_login_set_player_cnt(req, ack *common.NetPack) {
	svrId := req.ReadInt()
	cnt := req.ReadInt32()
	g_game_player_cnt.Store(svrId, cnt)
}
func GetFreeGameSvrId(version string) int {
	ret, minCnt := -1, int32(999999)
	ids, _ := meta.GetModuleIDs("game", version)
	for _, id := range ids {
		if v, ok := g_game_player_cnt.Load(id); ok {
			if cnt := v.(int32); cnt < minCnt {
				ret, minCnt = id, cnt
			}
		} else {
			ret, minCnt = id, 0
		}
	}
	return ret
}
