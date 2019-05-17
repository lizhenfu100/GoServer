/***********************************************************************
* @ 记录于账号上面的游戏信息
* @ brief
	1、参考QQ、微信做法，一套账号系统可关联多个游戏，复用社交数据

	2、须手选区服的游戏（同一账号可在各个区建多个角色），记录其登录服意义不大
		· 提供web接口供玩家查询，在哪些服有记录

* @ author zhoumf
* @ date 2018-3-20
***********************************************************************/
package gameInfo

import (
	"common"
)

type TGameInfo struct {
	LoginSvrId int //玩家所在区服（仅自动选服时有效，手动选服的游戏是0）
	GameSvrId  int

	JsonData string //各游戏独有的数据
}

func (self *TGameInfo) DataToBuf(buf *common.NetPack) {
	buf.WriteInt(self.LoginSvrId)
	buf.WriteInt(self.GameSvrId)
	buf.WriteString(self.JsonData)
}
func (self *TGameInfo) BufToData(buf *common.NetPack) {
	self.LoginSvrId = buf.ReadInt()
	self.GameSvrId = buf.ReadInt()
	if data := buf.ReadString(); data != "" {
		self.JsonData = data
	}
}

// game的分流节点，【须保证玩家分配到的节点不变】，不能动态增删
func ShuntGameSvr(gamelist []int, svrId *int, accountId uint32) bool {
	var ids []int
	for _, id := range gamelist {
		if id%10000 == *svrId%10000 { //svrId%10000相同，视为分流节点
			ids = append(ids, id)
		}
	}
	if length := uint32(len(ids)); length > 0 { //hash选择分流节点
		*svrId = ids[accountId%length]
		return true
	}
	return false
}
