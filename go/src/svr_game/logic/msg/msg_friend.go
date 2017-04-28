package msg

import (
	"common"

	"svr_game/logic/player"
)

func Rpc_Friend_Get_Info(req, ack *common.NetPack, ptr interface{}) {
	player := ptr.(*player.TPlayer)
	player.Friend.DataToBuf(ack)
}
func Rpc_Friend_Apply(req, ack *common.NetPack, ptr interface{}) {
	destPid := req.ReadUInt32()

	player.AsyncNotifyPlayer(destPid, func(destPtr *player.TPlayer) {
		player := ptr.(*player.TPlayer)
		destPtr.Friend.RecvApply(player.PlayerID, player.Name)
	})
}
func Rpc_Friend_Agree(req, ack *common.NetPack, ptr interface{}) {
	pos := req.ReadByte()

	player := ptr.(*player.TPlayer)
	player.Friend.Agree(pos)
}
func Rpc_Friend_Refuse(req, ack *common.NetPack, ptr interface{}) {
	pos := req.ReadByte()

	player := ptr.(*player.TPlayer)
	player.Friend.Refuse(pos)
}
func Rpc_Friend_Add(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()
	name := req.ReadString()

	player := ptr.(*player.TPlayer)
	player.Friend.AddFriend(pid, name)
}
func Rpc_Friend_Del(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*player.TPlayer)
	player.Friend.DelFriend(pid)
}
func Rpc_Friend_Add_Black(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()
	name := req.ReadString()

	player := ptr.(*player.TPlayer)
	player.Friend.AddBlack(pid, name)
}
func Rpc_Friend_Del_Black(req, ack *common.NetPack, ptr interface{}) {
	pid := req.ReadUInt32()

	player := ptr.(*player.TPlayer)
	player.Friend.DelBlack(pid)
}
