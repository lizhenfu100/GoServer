package player

import (
	"common"
)

// ------------------------------------------------------------
// -- 玩家常规显示数据
type TShowInfo struct {
	AccountId uint32
	Name      string
	Head      string
}

func (self *TShowInfo) BufToData(buf *common.NetPack) {
	self.AccountId = buf.ReadUInt32()
	self.Name = buf.ReadString()
	self.Head = buf.ReadString()
}
func (self *TShowInfo) DataToBuf(buf *common.NetPack) {
	buf.WriteUInt32(self.AccountId)
	buf.WriteString(self.Name)
	buf.WriteString(self.Head)
}
func (self *TPlayer) GetShowInfo() (ret TShowInfo) {
	ret.AccountId = self.AccountID
	ret.Name = self.Name
	ret.Head = self.Head
	return
}

// ------------------------------------------------------------
// -- rpc
func Rpc_game_get_show_info(req, ack *common.NetPack, this *TPlayer) {
	info := this.GetShowInfo()
	info.DataToBuf(ack)
}
