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
	账号下多角色：
		1、建表<Aid, 角色集>，登录先返回角色集合，再拉角色数据

* @ 分流节点（game连接同个db）
	、svrId%meta.KIdMod 即主节点id
	、id取模决定去哪个分流节点的
		· 节点不变时，玩家每次进的同个节点
		· 此点非常重要，game带状态的
		· 若每次可进不同节点，会状态错乱……比如延时写db，脏覆盖bug

* @ 逐级鉴权 Rpc_check_identity、Rpc_set_identity
	鉴权：验证是否能够使用某类服务，上下行参数各服务不同
	· 登录服鉴权，获取账号id、游戏服或网关地址
	· 网关鉴权，之后才能转发（可省去）
	· 游戏服鉴权，获取角色列表（可省去）
	· 角色鉴权，获取角色数据（可省去）

* @ 动态分流
	· game、center...连同个db的节点，可视为无状态的，能动态加
	· gateway：
		· 取模aid分流的，其它节点也会用
		· 单次登录中客户端与某个gateway绑定的
		· 这些强状态，导致不能随便加gateway（部署时就给够~囧）

* @ author zhoumf
* @ date 2018-3-19
***********************************************************************/
package logic

import (
	"common"
	"common/assert"
	"common/std/sign"
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

var g_token uint32

func Rpc_check_identity(req, ack *common.NetPack) { Rpc_login_account_login(req, ack) }
func Rpc_login_account_login(req, ack *common.NetPack) {
	version := req.ReadString()
	gameSvrId := 0
	if !conf.Auto_GameSvr {
		gameSvrId = req.ReadInt() //若手选区服则由客户端上报
	}
	gameName := req.ReadString()
	account := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString()
	sign.Decode(&account, &pwd)
	if !assert.IsDebug && gameName != conf.GameName {
		ack.WriteUInt16(err.LoginSvr_not_match)
		wechat.SendMsg("登录服不匹配：" + account)
	} else if centerId := netConfig.HashCenterID(account); centerId < 0 {
		ack.WriteUInt16(err.None_center_server)
	} else {
		//2、同步转至center验证账号信息，取得accountId、gameInfo
		if ok, _1 := accountLogin1(centerId, &gameSvrId, account, pwd, typ, ack); ok {
			//3、确定gameSvrId，处理gameInfo
			if e := accountLogin2(_1.AccountID, &gameSvrId, version, centerId); e == err.Success {
				//4、登录成功，设置各类信息，回复client待连接的节点(gateway或game)
				accountLogin3(&_1, gameSvrId, version, ack)
			} else {
				ack.WriteUInt16(e)
			}
		}
	}
}
func accountLogin1(centerId int, gameSvrId *int, account, pwd, typ string, ack *common.NetPack) (
	_ok bool,
	_1 gameInfo.TAccountClient) {
	if !cache.Get(account, pwd, ack) {
		netConfig.CallRpcCenter(centerId, enum.Rpc_center_account_login, func(buf *common.NetPack) {
			buf.WriteString(conf.GameName)
			buf.WriteString(account)
			buf.WriteString(pwd)
			buf.WriteString(typ)
			//buf.WriteString(conf.GameName) //Rpc_check_identity
		}, func(recvBuf *common.NetPack) {
			ack.WriteBuf(recvBuf.LeftBuf())
		})
		cache.Add(account, pwd, ack)
	}
	oldPos := ack.ReadPos //临时读取
	if ack.ReadUInt16() != err.Success {
		ack.ReadPos = oldPos
	} else {
		_1.BufToData(ack)
		loginId := ack.ReadInt()
		gameId := ack.ReadInt()
		ack.ClearBody()
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
func accountLogin2(aid uint32, gameSvrId *int, version string, centerId int) uint16 {
	gamelog.Debug("GameId:%v, version:%s", *gameSvrId, version)
	//选取gameSvrId：若账户未绑定游戏服，自动选取空闲节点，并绑定到账号上
	if *gameSvrId <= 0 {
		if *gameSvrId = GetFreeGameSvr(version); *gameSvrId <= 0 {
			return err.None_free_game_server
		} else if conf.Auto_GameSvr {
			errCode := err.None_center_server
			netConfig.CallRpcCenter(centerId, enum.Rpc_center_set_game_route2, func(buf *common.NetPack) {
				buf.WriteUInt32(aid)
				buf.WriteString(conf.GameName)
				buf.WriteInt(meta.G_Local.SvrID) //loginId
				buf.WriteInt(*gameSvrId)
			}, func(backBuf *common.NetPack) {
				errCode = backBuf.ReadUInt16()
				cache.Rpc_login_del_account_cache(backBuf, nil)
			})
			return errCode
		} else {
			return err.Success
		}
	} else if p := meta.ShuntSvr(gameSvrId, meta.GetMetas("game", version), aid); p == nil {
		return err.None_free_game_server
	} else {
		gamelog.Debug("Login game svrId: %d", *gameSvrId)
		return err.Success
	}
}
func accountLogin3(pInfo *gameInfo.TAccountClient, gameSvrId int, version string, ack *common.NetPack) {
	//生成一个临时token，发给gamesvr、client用以登录验证
	token, errCode := atomic.AddUint32(&g_token, 1), err.Success
	if p, ok := netConfig.GetGameRpc(gameSvrId); ok {
		p.CallRpcSafe(enum.Rpc_set_identity, func(buf *common.NetPack) {
			buf.WriteUInt32(token)
			buf.WriteUInt32(pInfo.AccountID)
		}, func(backBuf *common.NetPack) {
			cnt := backBuf.ReadInt32()
			errCode = backBuf.ReadUInt16()
			g_game_player_cnt.Store(gameSvrId, cnt)
		})
	}
	//回复client待连接的节点(gateway或game)
	gatewayId := netConfig.HashGatewayID(pInfo.AccountID)
	var pMetaToClient *meta.Meta
	if pMetaToClient = meta.GetMeta("gateway", gatewayId); pMetaToClient != nil {
		if p, ok := netConfig.GetGatewayRpc(gatewayId); ok {
			p.CallRpcSafe(enum.Rpc_set_identity, func(buf *common.NetPack) {
				buf.WriteUInt32(token)
				buf.WriteUInt32(pInfo.AccountID)
				buf.WriteInt(gameSvrId)
			}, nil)
		}
	} else if pMetaToClient = meta.GetMeta("game", gameSvrId); pMetaToClient == nil {
		errCode = err.None_game_server
	}
	if conf.GameName == "SoulKnight" { //TODO:待删除，老包无gateway
		if common.CompareVersion(version, "2.6.0") < 0 {
			if pMetaToClient = meta.GetMeta("game", gameSvrId); pMetaToClient == nil {
				errCode = err.None_game_server
			}
		}
	}
	if ack.WriteUInt16(errCode); pMetaToClient != nil { //回复client要连接的目标节点
		ack.WriteUInt32(pInfo.AccountID)
		ack.WriteString(pMetaToClient.OutIP)
		ack.WriteUInt16(pMetaToClient.Port())
		ack.WriteUInt32(token)
		ack.WriteUInt8(pInfo.IsValidEmail)
	}
}

// ------------------------------------------------------------
// 转发至center
func Rpc_login_to_center_by_str(req, ack *common.NetPack) {
	rpcId := req.ReadUInt16() //目标rpc
	rpcData := req.LeftBuf()
	str := req.ReadString() //accountName/bindVal
	sign.Decode(&str)
	centerId := netConfig.HashCenterID(str)
	switch rpcId {
	case enum.Rpc_center_reg_if, enum.Rpc_center_reg_if2:
		if cache.IsExist(str) {
			ack.WriteUInt16(err.Success)
			return
		}
	case enum.Rpc_center_bind_info, enum.Rpc_center_bind_info2:
		cache.Del(str)
	}
	netConfig.CallRpcCenter(centerId, rpcId, func(buf *common.NetPack) {
		buf.WriteBuf(rpcData)
	}, func(recvBuf *common.NetPack) {
		ack.WriteBuf(recvBuf.LeftBuf())
	})
	gamelog.Track("relay center: %d, %s, %d", rpcId, str, centerId)
}

// ------------------------------------------------------------
// 大区游戏服列表
func Rpc_login_get_game_list(req, ack *common.NetPack) {
	version := req.ReadString()
	list := meta.GetMetas("game", version)
	ack.WriteByte(byte(len(list)))
	for _, p := range list {
		ack.WriteInt(p.SvrID)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())
		ack.WriteString(p.SvrName)
		if v, ok := g_game_player_cnt.Load(p.SvrID); ok {
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
	for _, p := range meta.GetMetas("game", version) {
		if v, ok := g_game_player_cnt.Load(p.SvrID); ok {
			if cnt := v.(int32); cnt < minCnt {
				ret, minCnt = p.SvrID, cnt
			}
		} else {
			ret, minCnt = p.SvrID, 0
		}
	}
	return ret
}
