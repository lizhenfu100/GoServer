/***********************************************************************
* @ 战斗匹配
* @ brief
	*、玩法类型
		1、球球类型的，允许中途加入房间，匹配得送到战斗服做
		2、王者这样的，匹配直接在GameSvr做，组好一场人，全送进某个战斗服即可

* @ author zhoumf
* @ date 2017-6-5
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

func Rpc_Battle_Begin(req, ack *common.NetPack, ptr interface{}) {
	battle := common.NewNetPackCap(256)
	cnt := req.ReadByte()
	battle.WriteByte(cnt)
	for i := byte(0); i < cnt; i++ {
		pid := req.ReadUInt32()
		if player := _FindPlayerInCache(pid); player != nil {
			// pack player battle data
			battle.WriteUInt32(pid)
		} else {
			ack.WriteInt8(-1)
			return
		}
	}
	battle.SetRpc("relay_battle_data")
	api.SendToCross(battle)
	ack.WriteInt8(1)
}
