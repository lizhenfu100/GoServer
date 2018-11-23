/***********************************************************************
* @ 玩家登录
* @ brief
	1、验证相关，放登录服，减缓游戏服压力，比如：
		、每个账号在同一区服，只允许建一个或几个角色
		、角色名不能跟别人重复（要查找整张表呀）

	2、登录服验证通过后，将玩家数据载入DBCache(比如redis)，下发AccountId/PlayerId之类的给Client

	3、Client再到游戏服真正登录时，就只需从DBCache载入了

* @ author zhoumf
* @ date 2017-4-26
***********************************************************************/
package player

import (
	"common"
	"common/format"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
	"sync"
	"sync/atomic"
	"tcp"
)

// -------------------------------------
// -- 玩家登录
// Notice：登录、创建角色，可做成普通rpc，用以建立玩家缓存
func Rpc_game_login(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()

	// game直连了login，须校验登录token
	for _, v := range netConfig.G_Local_Meta.ConnectLst {
		if v == "login" {
			token := req.ReadUInt32()
			if CheckLoginToken(accountId, token) {
				_NotifyPlayerCnt(netConfig.G_Local_Meta.SvrID, g_player_cnt)
			} else {
				ack.WriteUInt16(err.Token_verify_err)
				return
			}
			break
		}
	}
	//TODO:zhoumf: 读数据库同步的，比较耗时，直接读的方式不适合外网；可转入线程池再通知回主线程
	if this := FindWithDB_AccountId(accountId); this == nil {
		ack.WriteUInt16(err.Account_have_none_player) //notify client to create new player
	} else {
		this.Login(conn)
		gamelog.Debug("Player Login: %s, accountId(%d)", this.Name, this.PlayerID)
		ack.WriteUInt16(err.Success)
		ack.WriteUInt32(this.PlayerID)
		ack.WriteString(this.Name)
	}
}
func Rpc_game_create_player(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	playerName := req.ReadString()

	if !format.CheckName(playerName) { //名字不合格
		ack.WriteUInt32(0)
	} else if this := NewPlayerInDB(accountId, playerName); this == nil {
		ack.WriteUInt32(0)
	} else {
		this.Login(conn)
		gamelog.Debug("Create NewPlayer: accountId(%d) name(%s)", accountId, playerName)
		ack.WriteUInt32(this.PlayerID)
	}
}
func Rpc_game_logout(req, ack *common.NetPack, this *TPlayer) {
	this.Logout()
}

func Rpc_game_get_player_cnt(req, ack *common.NetPack, conn *tcp.TCPConn) {
	cnt := atomic.LoadInt32(&g_player_cnt)
	ack.WriteInt32(cnt)
}

// -------------------------------------
// -- 后台账号验证
var g_login_token sync.Map //<accountId, token>

func Rpc_game_login_token(req, ack *common.NetPack, conn *tcp.TCPConn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	g_login_token.Store(accountId, token)

	cnt := atomic.LoadInt32(&g_player_cnt)
	ack.WriteInt32(cnt)
}
func CheckLoginToken(accountId, token uint32) bool {
	if value, ok := g_login_token.Load(accountId); ok {
		return token == value
	}
	return false
}

// ------------------------------------------------------------
// -- 游戏服在线人数
func _NotifyPlayerCnt(gameSvrId int, cnt int32) {
	ids, _ := meta.GetModuleIDs("login", netConfig.G_Local_Meta.Version)
	for _, id := range ids {
		if addr := netConfig.GetHttpAddr("login", id); addr != "" {
			http.CallRpc(addr, enum.Rpc_login_set_player_cnt, func(buf *common.NetPack) {
				buf.WriteInt(gameSvrId)
				buf.WriteInt32(cnt)
			}, nil)
		}
	}
}
