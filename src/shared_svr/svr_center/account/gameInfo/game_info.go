/***********************************************************************
* @ 记录于账号上面的游戏信息
* @ brief
	1、游戏的角色id，暗含了大区信息；头两字节分别是loginId、gameId

	2、手选区服的游戏（同一账号可在各个区建多个角色），记录其登录服意义不大
		· 提供web接口供玩家查询，在哪些服有记录

* @ author zhoumf
* @ date 2018-3-20
***********************************************************************/
package gameInfo

import "common"

type TGameInfo struct {
	LoginSvrId int //玩家所在区服（仅自动选服时有效，手动选服的游戏是0）
	GameSvrId  int
	JsonData   string //各游戏独有的数据
}

// ------------------------------------------------------------
// 打包给客户端的账号信息
type TAccountClient struct {
	AccountID    uint32
	IsValidEmail uint8
}

func (self *TAccountClient) DataToBuf(buf *common.NetPack) {
	buf.WriteUInt32(self.AccountID)
	buf.WriteUInt8(self.IsValidEmail)
}
func (self *TAccountClient) BufToData(buf *common.NetPack) {
	self.AccountID = buf.ReadUInt32()
	self.IsValidEmail = buf.ReadUInt8()
}
