package player

import (
	"common"
)

func Rpc_Player_Login(req, ack *common.NetPack) {
	//req: accountId, loginKey(账号服生成的登录验证码)
	//ack: playerId, data
	accountId := req.ReadUint32()

	if player := FindWithDB_AccountId(accountId); player != nil {

	} else {
		//notify client to create new player
	}
}
func Rpc_Player_Logout(req, ack *common.NetPack) {
	//req: playerId
	playerId := req.ReadUint32()
	DelPlayerCache(playerId)
}
func Rpc_Player_Create(req, ack *common.NetPack) {
	//req: accountId, loginKey, playerName
	//ack: playerId
	accountId := req.ReadUint32()
	playerName := req.ReadString()
	if player := AddNewPlayer(accountId, playerName); player != nil {
		ack.WriteUint32(player.Base.PlayerID)
	}
}
