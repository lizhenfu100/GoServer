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
)

// -------------------------------------
// -- 玩家登录
func Rpc_game_login(req, ack *common.NetPack) {
	//req: accountId, token(账号服生成的登录验证码)
	//ack: playerId, data
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if !CheckLoginToken(accountId, token) {
		ack.WriteInt8(-1)
	} else {
		player := FindWithDB_AccountId(accountId)
		if player == nil {
			ack.WriteInt8(-2) //notify client to create new player
		} else {
			player.Login()
			gamelog.Debug("Player Login: %s, pid(%d), accountId(%d)", player.Name, player.PlayerID, player.AccountID)
			ack.WriteInt8(1)
			ack.WriteUInt32(player.PlayerID)
			ack.WriteString(player.Name)
		}
	}
}
func Rpc_game_logout(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*TPlayer)
	player.Logout()
}
func Rpc_game_player_create(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()
	playerName := req.ReadString()

	if !CheckLoginToken(accountId, token) { //token验证
		ack.WriteUInt32(0)
	} else if !format.CheckName(playerName) { //名字不合格
		ack.WriteUInt32(0)
	} else if player := AddNewPlayer(accountId, playerName); player != nil {
		gamelog.Debug("Create NewPlayer: accountId(%d) name(%s) pid(%d)",
			accountId, playerName, player.PlayerID)
		player.Login()
		ack.WriteUInt32(player.PlayerID)
	} else {
		ack.WriteUInt32(0)
	}
}
func Rpc_game_heart_beat(req, ack *common.NetPack, ptr interface{}) {
}

// -------------------------------------
// -- 后台账号验证
var g_login_token sync.Map

func Rpc_game_login_token(req, ack *common.NetPack) {
	id := req.ReadUInt32()
	token := req.ReadUInt32()
	g_login_token.Store(id, token)
}
func CheckLoginToken(id, token uint32) bool {
	if value, ok := g_login_token.Load(id); ok {
		return token == value
	}
	return false
}
