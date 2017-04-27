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
)

func Rpc_Player_Login(req, ack *common.NetPack, ptr interface{}) {
	//req: accountId, loginKey(账号服生成的登录验证码)
	//ack: playerId, data
	accountId := req.ReadUInt32()

	player := FindWithDB_AccountId(accountId)

	if player != nil {
		player.OnLogin()
		fmt.Println(player)
	} else {
		//notify client to create new player
	}
}
func Rpc_Player_Logout(req, ack *common.NetPack, ptr interface{}) {

	player := ptr.(*TPlayer)

	player.OnLogin()
}
func Rpc_Player_Create(req, ack *common.NetPack, ptr interface{}) {
	//req: accountId, loginKey, playerName
	//ack: playerId
	accountId := req.ReadUInt32()
	playerName := req.ReadString()
	player := AddNewPlayer(accountId, playerName)
	if player != nil {
		gamelog.Info("Create New Player: %s(%d)", playerName, player.PlayerID)
		ack.WriteUInt32(player.PlayerID)
	}
}
