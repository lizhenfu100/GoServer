/***********************************************************************
* @ 玩家登录
* @ brief
	1、验证相关，可放登录服(目前在Center)做，减缓游戏服压力，比如：
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
	"fmt"
	"gamelog"
	"netConfig"
	"svr_game/api"
	"svr_game/center"
)

func Rpc_Player_Login(req, ack *common.NetPack) {
	//req: accountId, token(账号服生成的登录验证码)
	//ack: playerId, data
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if !center.CheckLoginToken(accountId, token) {
		ack.WriteInt8(-1)
	} else {
		player := FindWithDB_AccountId(accountId)
		if player == nil {
			ack.WriteInt8(-2) //notify client to create new player
		} else {
			player.Login()
			fmt.Println("Player_Login:\n", player)
			ack.WriteInt8(1)
			ack.WriteUInt32(player.PlayerID)
			ack.WriteString(player.Name)

			// notify svr_center login success
			api.CallRpcCenter("rpc_center_login_game_success", func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteUInt32(uint32(netConfig.G_Local_SvrID))
			}, nil)
		}
	}
}
func Rpc_Player_Logout(req, ack *common.NetPack, ptr interface{}) {

	player := ptr.(*TPlayer)

	player.Logout()
}
func Rpc_Player_Create(req, ack *common.NetPack) {
	//req: accountId, playerName
	//ack: playerId
	accountId := req.ReadUInt32()
	playerName := req.ReadString()

	if player := AddNewPlayer(accountId, playerName); player != nil {
		gamelog.Info("Create NewPlayer: accountId(%d) name(%s) pid(%d)",
			accountId, playerName, player.PlayerID)
		player.Login()
		ack.WriteUInt32(player.PlayerID)
	} else {
		ack.WriteUInt32(0)
	}
}
func Rpc_Heart_Beat(req, ack *common.NetPack, ptr interface{}) {
}
