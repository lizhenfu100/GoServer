package player

import (
	"common"
	"fmt"
	"gamelog"
)

func Rpc_Player_Login(req, ack *common.ByteBuffer) interface{} {
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
	return player
}
func Rpc_Player_Logout(req, ack *common.ByteBuffer) interface{} {
	//req: playerId
	playerId := req.ReadUInt32()
	player := FindPlayerInCache(playerId)
	if player != nil {
		player.OnLogin()
	}
	return player
}
func Rpc_Player_Create(req, ack *common.ByteBuffer) interface{} {
	//req: accountId, loginKey, playerName
	//ack: playerId
	accountId := req.ReadUInt32()
	playerName := req.ReadString()
	player := AddNewPlayer(accountId, playerName)
	if player != nil {
		gamelog.Info("Create New Player: %s(%d)", playerName, player.PlayerID)
		ack.WriteUInt32(player.PlayerID)
	}
	return player
}
