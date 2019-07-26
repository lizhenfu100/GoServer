/***********************************************************************
* @ game svr list
* @ brief
	客户端选取登录服：
		1、客户端显示大区列表，玩家自己选择
		2、login iplist 写成前端配置，玩家注册时挨个ping，取延时最小的去注册
	玩家手选区服：
		1、下发区服列表，client选定具体gamesvr节点登录
		2、client本地存储gamesvr节点编号，登录时上报
	后台自动分服：
		1、验证账号，若账号中无绑定区服信息，进入创建角色流程
		2、选择最空闲gamesvr节点，创建角色成功，将区服编号绑账号上
		3、下次登录，验证账号后即定位具体gamesvr节点

* @ 分流节点（game连接同个db）
	、常规节点的svrId四位数以内，分流节点是svrId+10000
	、svrId%10000 即入库的节点id
	、id取模决定去哪个分流节点的
		· 节点不变时，玩家每次进的同个节点
		· 此点非常重要，game带状态的
		· 若每次可进不同节点，会状态错乱……比如延时写db，脏覆盖bug

* @ 动态分流
	、无zookeeper的架构，只能重启扩容
	、取模方式的，不能动态分流，如：gateway、friend、game

* @ author zhoumf
* @ date 2018-3-19
***********************************************************************/
package logic

import (
	"common"
	"common/tool/wechat"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	info "shared_svr/svr_center/account/gameInfo"
	"sync"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) {
	version := req.ReadString()
	gameSvrId := 0
	if conf.HandPick_GameSvr {
		gameSvrId = req.ReadInt() //若手选区服则由客户端上报
	}
	//1、临时读取buffer数据，将数据转发center
	oldPos := req.ReadPos
	gameName := req.ReadString()
	account := req.ReadString()
	req.ReadPos = oldPos

	if gameName != conf.GameName {
		ack.WriteUInt16(err.LoginSvr_not_match)
		wechat.SendMsg("LoginSvr_not_match: " + conf.GameName + account)
	} else {
		//2、同步转至center验证账号信息，取得accountId、gameInfo
		centerSvrId := netConfig.HashCenterID(account)
		if ok, _1, _2 := accountLogin1(centerSvrId, req, ack); ok {
			//3、确定gameSvrId，处理gameInfo
			errCode := accountLogin2(_1.AccountID, &gameSvrId, &_2, version, gameName, centerSvrId)
			if errCode == err.Success {
				//4、登录成功，设置各类信息，回复client待连接的节点(gateway或game)
				accountLogin3(&_1, gameSvrId, ack)
			} else {
				ack.WriteUInt16(errCode)
			}
		}
	}
}
func accountLogin1(centerSvrId int, req, ack *common.NetPack) (
	_ok bool,
	_1 info.TAccountClient,
	_2 info.TGameInfo) {
	//同步至center验证账号信息，取得accountId、gameInfo
	netConfig.SyncRelayToCenter(centerSvrId, enum.Rpc_center_account_login, req, ack)

	//临时读取buffer数据
	oldPos := ack.ReadPos
	if ack.ReadUInt16() != err.Success {
		ack.ReadPos = oldPos
		_ok = false
	} else {
		_1.BufToData(ack)
		_2.BufToData(ack)
		ack.Clear()

		//Notice: 玩家不是这个大区的，更换大区再登录
		if _2.LoginSvrId > 0 && _2.LoginSvrId != meta.G_Local.SvrID {
			ack.WriteUInt16(err.LoginSvr_not_match)
			_ok = false
		} else {
			_ok = true
		}
	}
	return
}
func accountLogin2(
	accountId uint32,
	gameSvrId *int,
	pInfo *info.TGameInfo,
	version, gameName string,
	centerSvrId int) (errCode uint16) {
	gamelog.Track("GameInfo:%v, version:%s gameName:%s", pInfo, version, gameName)
	if !conf.HandPick_GameSvr {
		//选取gameSvrId：若账户未绑定游戏服，自动选取空闲节点，并绑定到账号上
		if *gameSvrId = pInfo.GameSvrId; *gameSvrId <= 0 {
			if *gameSvrId = GetFreeGameSvr(version); *gameSvrId <= 0 {
				errCode = err.None_free_game_server
			} else {
				errCode = err.None_center_server
				netConfig.CallRpcCenter(centerSvrId, enum.Rpc_center_set_game_info,
					func(buf *common.NetPack) {
						buf.WriteUInt32(accountId)
						buf.WriteString(gameName)
						pInfo.GameSvrId = *gameSvrId % 10000
						pInfo.LoginSvrId = meta.G_Local.SvrID % 10000
						pInfo.DataToBuf(buf)
					}, func(backBuf *common.NetPack) {
						errCode = backBuf.ReadUInt16()
					})
			}
			return
		}
	}
	if !info.ShuntSvr(meta.GetModuleIDs("game", version), gameSvrId, accountId) {
		errCode = err.None_free_game_server
	} else {
		errCode = err.Success
		gamelog.Track("Login game svrId: %d", *gameSvrId)
	}
	return
}
func accountLogin3(pInfo *info.TAccountClient, gameSvrId int, ack *common.NetPack) {
	//登录成功，回复client待连接的节点(gateway或game)
	gatewayId := netConfig.HashGatewayID(pInfo.AccountID) //此玩家要连接的gateway
	errCode := err.Unknow_error

	//生成一个临时token，发给gamesvr、client用以登录验证
	token := atomic.AddUint32(&g_login_token, 1)

	var pMetaToClient *meta.Meta //回复client要连接的目标节点
	//【Notice: tcp.CallRpc接口不是线程安全的，http后台不适用】
	if conn := netConfig.GetGatewayConn(gatewayId); conn != nil {
		conn.CallRpcSafe(enum.Rpc_gateway_login_token, func(buf *common.NetPack) {
			buf.WriteUInt32(token)
			buf.WriteUInt32(pInfo.AccountID)
			buf.WriteInt(gameSvrId)
		}, func(backBuf *common.NetPack) {
			cnt := backBuf.ReadInt32()
			g_game_player_cnt.Store(gameSvrId, cnt)
		})
		pMetaToClient = meta.GetMeta("gateway", gatewayId)
		errCode = err.None_gateway
	} else {
		netConfig.CallRpcGame(gameSvrId, enum.Rpc_game_login_token, func(buf *common.NetPack) {
			buf.WriteUInt32(token)
			buf.WriteUInt32(pInfo.AccountID)
		}, func(backBuf *common.NetPack) {
			cnt := backBuf.ReadInt32()
			g_game_player_cnt.Store(gameSvrId, cnt)
		})
		pMetaToClient = meta.GetMeta("game", gameSvrId)
		errCode = err.None_game_server
	}
	if pMetaToClient != nil { //回复client要连接的目标节点
		ack.WriteUInt16(err.Success)
		ack.WriteUInt32(pInfo.AccountID)
		ack.WriteString(pMetaToClient.OutIP)
		ack.WriteUInt16(pMetaToClient.Port())
		ack.WriteUInt32(token)
		ack.WriteUInt8(pInfo.IsValidEmail)
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
func Rpc_login_get_game_list(req, ack *common.NetPack) {
	version := req.ReadString()

	ids := meta.GetModuleIDs("game", version)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		p := meta.GetMeta("game", id)
		ack.WriteInt(id)
		ack.WriteString(p.SvrName)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())

		if v, ok := g_game_player_cnt.Load(id); ok {
			ack.WriteInt32(v.(int32))
		} else {
			ack.WriteInt32(0)
		}
	}
}

//【维护不同进程的数据一致，须保证“增删改查”，成本过高】
// 登录流程中附带，或，各节点主动周期性上报
var g_game_player_cnt sync.Map //<gameSvrId, playerCnt>

func GetFreeGameSvr(version string) int {
	ret, minCnt := -1, int32(999999)
	ids := meta.GetModuleIDs("game", version)
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
