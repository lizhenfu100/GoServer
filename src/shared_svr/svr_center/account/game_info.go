/***********************************************************************
* @ 记录于账号上面的游戏信息
* @ brief
	1、参考QQ、微信做法，一套账号系统可关联多个游戏，复用社交数据

* @ author zhoumf
* @ date 2018-3-20
***********************************************************************/
package account

import (
	"common"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type IGameInfo interface {
	DataToBuf(*common.NetPack)
	BufToData(*common.NetPack)
}

func Rpc_center_set_game_info(req, ack *common.NetPack) {
	accountId := req.ReadUInt32()
	gameName := req.ReadString()

	if ptr := GetAccountInCache(accountId); ptr != nil {
		if info := ptr.GetGameInfo(gameName); info != nil {
			info.BufToData(req)
			dbmgo.UpdateToDB("Account", bson.M{"_id": accountId}, bson.M{"$set": bson.M{strings.ToLower(gameName): info}})
		}
	}
}
