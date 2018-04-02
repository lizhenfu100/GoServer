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
	"sync"
	"tcp"
)

// -------------------------------------
// -- 玩家登录
// Notice：登录、创建角色，可做成普通rpc，用以建立玩家缓存
func Rpc_game_login(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()

	//TODO:zhoumf: 读数据库同步的，比较耗时，直接读的方式不适合外网
	if player := FindWithDB_AccountId(accountId); player == nil {
		ack.WriteInt8(-2) //notify client to create new player
	} else {
		player.Login(conn)
		gamelog.Debug("Player Login: %s, accountId(%d)", player.Name, player.AccountID)
		ack.WriteInt8(1)
		ack.WriteUInt32(player.AccountID)
		ack.WriteString(player.Name)
	}
}
func Rpc_game_create_player(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	playerName := req.ReadString()

	if !format.CheckName(playerName) { //名字不合格
		ack.WriteUInt32(0)
	} else if player := NewPlayerInDB(accountId, playerName); player == nil {
		ack.WriteUInt32(0)
	} else {
		gamelog.Debug("Create NewPlayer: accountId(%d) name(%s)", accountId, playerName)
		player.Login(conn)
		ack.WriteUInt32(player.AccountID)
	}
}
func Rpc_game_logout(req, ack *common.NetPack, player *TPlayer) {
	player.Logout()
}

func Rpc_game_get_player_cnt(req, ack *common.NetPack, conn *tcp.TCPConn) {
	ack.WriteInt32(g_player_cnt)
}

// -------------------------------------
// -- 后台账号验证
var g_login_token sync.Map

func Rpc_game_login_token(req, ack *common.NetPack, conn *tcp.TCPConn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	g_login_token.Store(accountId, token)

	ack.WriteInt32(g_player_cnt)
}
func CheckLoginToken(id, token uint32) bool {
	if value, ok := g_login_token.Load(id); ok {
		return token == value
	}
	return false
}
