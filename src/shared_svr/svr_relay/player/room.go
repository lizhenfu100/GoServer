package player

import (
	"common"
	"gamelog"
	"generate_out/rpc/enum"
)

type Room struct {
	list []*Player
}

func (self *Player) NewRoom() *Room {
	r := new(Room)
	r.list = append(r.list, self)
	self.room = r
	return r
}
func (self *Player) JoinRoom(ownerId uint32) {
	if owner := FindPlayer(ownerId); owner != nil && owner.room != nil {
		if owner.room.Add(self) {
			owner.Conn.CallRpc(enum.Rpc_client_on_other_join_team, func(buf *common.NetPack) {
				buf.WriteUInt32(self.Pid)
			}, nil) //通知主机
		} else {
			gamelog.Error("add room: %d", self.Pid)
		}
	} else {
		gamelog.Error("none room: %d", ownerId)
	}
}
func (self *Player) ExitRoom() {
	if room := self.room; room != nil {
		isHost := room.list[0] == self
		if room.Del(self); len(room.list) > 0 {
			if isHost { //房主退出，广播给成员们
				for _, v := range room.list {
					v.Conn.CallRpc(enum.Rpc_client_on_exit_team, func(buf *common.NetPack) {
						buf.WriteUInt32(self.Pid)
					}, nil)
				}
				//TODO:zhoumf:广播销毁房间
			} else { //成员退出，通知房主
				room.list[0].Conn.CallRpc(enum.Rpc_client_on_exit_team, func(buf *common.NetPack) {
					buf.WriteUInt32(self.Pid)
				}, nil)
			}
		}
		//通知自己的客户端
		self.Conn.CallRpc(enum.Rpc_client_on_exit_team, func(buf *common.NetPack) {
			buf.WriteUInt32(self.Pid)
		}, nil)
	}
}
func (self *Room) Have(p *Player) bool {
	for _, v := range self.list {
		if v == p {
			return true
		}
	}
	return false
}
func (self *Room) Add(p *Player) bool {
	if !self.Have(p) {
		self.list = append(self.list, p)
		p.room = self
		return true
	}
	return false
}
func (self *Room) Del(p *Player) {
	for i := len(self.list) - 1; i >= 0; i-- {
		if v := self.list[i]; v == p {
			self.list = append(self.list[:i], self.list[i+1:]...)
			p.room = nil
		}
	}
}
