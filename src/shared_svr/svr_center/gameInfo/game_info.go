/***********************************************************************
* @ 记录于账号上面的游戏信息
* @ brief
	1、参考QQ、微信做法，一套账号系统可关联多个游戏，复用社交数据

* @ author zhoumf
* @ date 2018-3-20
***********************************************************************/
package gameInfo

import "common"

type TGameInfo struct {
	SvrId int //玩家所在区服

	JsonData string //各游戏独有的数据
}

func (self *TGameInfo) DataToBuf(buf *common.NetPack) {
	buf.WriteInt(self.SvrId)
	buf.WriteString(self.JsonData)
}
func (self *TGameInfo) BufToData(buf *common.NetPack) {
	self.SvrId = buf.ReadInt()
	self.JsonData = buf.ReadString()
}
