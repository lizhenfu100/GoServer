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
	"common/assert"
	"common/tool/wechat"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/account/gameInfo"
	"shared_svr/svr_login/logic/cache"
	"sync"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) {
	version := req.ReadString()
	gameSvrId := 0
	if !conf.Auto_GameSvr {
		gameSvrId = req.ReadInt() //若手选区服则由客户端上报
	}
	//1、临时读取buffer数据，将数据转发center
	oldPos := req.ReadPos
	gameName := req.ReadString()
	account := req.ReadString()
	req.ReadPos = oldPos

	if !assert.IsDebug && gameName != conf.GameName {
		ack.WriteUInt16(err.LoginSvr_not_match)
		wechat.SendMsg("登录服不匹配：" + conf.GameName + account)
	} else if centerId := netConfig.HashCenterID(account); centerId < 0 {
		ack.WriteUInt16(err.None_center_server)
	} else {
		//2、同步转至center验证账号信息，取得accountId、gameInfo
		if ok, _1 := accountLogin1(centerId, &gameSvrId, req, ack); ok {
			//3、确定gameSvrId，处理gameInfo
			errCode := accountLogin2(_1.AccountID, &gameSvrId, version, gameName, centerId)
			if errCode == err.Success {
				//4、登录成功，设置各类信息，回复client待连接的节点(gateway或game)
				accountLogin3(&_1, gameSvrId, ack)
			} else {
				ack.WriteUInt16(errCode)
			}
		}
	}
}
func accountLogin1(centerId int, gameSvrId *int, req, ack *common.NetPack) (
	_ok bool,
	_1 gameInfo.TAccountClient) {
	if ok, key, pwd := cache.AccountLogin(req, ack); !ok {
		//同步至center验证账号信息，取得accountId、gameInfo
		netConfig.SyncRelayToCenter(centerId, enum.Rpc_center_account_login, req, ack)
		cache.Add(key, pwd, ack)
	}
	oldPos := ack.ReadPos //临时读取
	if ack.ReadUInt16() != err.Success {
		ack.ReadPos = oldPos
	} else {
		_1.BufToData(ack)
		loginId := ack.ReadInt()
		gameId := ack.ReadInt()
		ack.Clear()
		_ok = true
		if conf.Auto_GameSvr {
			if *gameSvrId = gameId; loginId > 0 && loginId != meta.G_Local.SvrID {
				ack.WriteUInt16(err.LoginSvr_not_match)
				_ok = false
			}
		}
	}
	return
}
func accountLogin2(aid uint32, gameSvrId *int, version, gameName string, centerId int) (errCode uint16) {
	gamelog.Debug("GameId:%v, version:%s", *gameSvrId, version)
	//选取gameSvrId：若账户未绑定游戏服，自动选取空闲节点，并绑定到账号上
	if *gameSvrId <= 0 {
		if *gameSvrId = GetFreeGameSvr(version); *gameSvrId <= 0 {
			errCode = err.None_free_game_server
		} else if conf.Auto_GameSvr {
			errCode = err.None_center_server
			netConfig.CallRpcCenter(centerId, enum.Rpc_center_set_game_route, func(buf *common.NetPack) {
				buf.WriteUInt32(aid)
				buf.WriteString(gameName)
				buf.WriteInt(meta.G_Local.SvrID % 10000) //loginId
				buf.WriteInt(*gameSvrId % 10000)
			}, func(backBuf *common.NetPack) {
				errCode = backBuf.ReadUInt16()
				cache.Rpc_login_del_account_cache(backBuf, nil)
			})
		}
	} else if !gameInfo.ShuntSvr(meta.GetModuleIDs("game", version), gameSvrId, aid) {
		errCode = err.None_free_game_server
	} else {
		errCode = err.Success
		gamelog.Debug("Login game svrId: %d", *gameSvrId)
	}
	return
}
func accountLogin3(pInfo *gameInfo.TAccountClient, gameSvrId int, ack *common.NetPack) {
	//生成一个临时token，发给gamesvr、client用以登录验证
	token := atomic.AddUint32(&g_login_token, 1)

	if p, ok := netConfig.GetGameRpc(gameSvrId); ok {
		p.CallRpcSafe(enum.Rpc_game_login_token, func(buf *common.NetPack) {
			buf.WriteUInt32(token)
			buf.WriteUInt32(pInfo.AccountID)
		}, func(backBuf *common.NetPack) {
			cnt := backBuf.ReadInt32()
			g_game_player_cnt.Store(gameSvrId, cnt)
		})
	}

	//回复client待连接的节点(gateway或game)
	var pMetaToClient *meta.Meta
	errCode := err.Unknow_error
	gatewayId := netConfig.HashGatewayID(pInfo.AccountID)
	if pMetaToClient = meta.GetMeta("gateway", gatewayId); pMetaToClient != nil {
		if p, ok := netConfig.GetGatewayRpc(gatewayId); ok {
			p.CallRpcSafe(enum.Rpc_gateway_login_token, func(buf *common.NetPack) {
				buf.WriteUInt32(token)
				buf.WriteUInt32(pInfo.AccountID)
				buf.WriteInt(gameSvrId)
			}, nil)
		}
	} else {
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

	if rpcId == enum.Rpc_center_reg_if { //渠道号登录都会先调这个，蛋疼~坑啊
		if cache.IsExist(strKey) {
			ack.WriteUInt16(err.Success)
			return
		}
	}
	svrId := netConfig.HashCenterID(strKey)
	netConfig.SyncRelayToCenter(svrId, rpcId, req, ack)
	gamelog.Track("relay center: %d, %s, %d", rpcId, strKey, svrId)
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

// 维护不同进程的数据一致，须保证“增删改查”，成本过高
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
