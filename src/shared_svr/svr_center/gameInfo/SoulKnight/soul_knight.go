package SoulKnight //元气骑士

import "common"

type TGameInfo struct {
	SvrId int //玩家所在区服
}

func (self *TGameInfo) DataToBuf(buf *common.NetPack) {
	buf.WriteInt(self.SvrId)
}
func (self *TGameInfo) BufToData(buf *common.NetPack) {
	self.SvrId = buf.ReadInt()
}
