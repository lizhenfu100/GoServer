package player

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
)

type TFriendModule struct {
	owner *TPlayer
}

// ------------------------------------------------------------
// -- 框架接口
func (self *TFriendModule) InitAndInsert(p *TPlayer) { self.owner = p }
func (self *TFriendModule) LoadFromDB(p *TPlayer)    { self.owner = p }
func (self *TFriendModule) WriteToDB()               {}
func (self *TFriendModule) OnLogin() {
	netConfig.CallRpcFriend(self.owner.AccountID, enum.Rpc_friend_get_friend_list, func(buf *common.NetPack) {
	}, func(recvBuf *common.NetPack) {
		cnt := recvBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			friendId := recvBuf.ReadUInt32()

			//通告好友我上线了
			netConfig.CallRpcGateway(friendId, enum.Rpc_client_friend_login, func(buf *common.NetPack) {
				buf.WriteUInt32(self.owner.AccountID)
			}, nil)

			//收集在线好友信息
			netConfig.CallRpcGateway(friendId, enum.Rpc_game_get_show_info, func(buf *common.NetPack) {
			}, func(recvBuf *common.NetPack) {
				netConfig.CallRpcGateway(self.owner.AccountID, enum.Rpc_client_friend_add, func(buf *common.NetPack) {
					buf.WriteBuf(recvBuf.LeftBuf())
				}, nil)
			})
		}
	})
}
func (self *TFriendModule) OnLogout() {
	netConfig.CallRpcFriend(self.owner.AccountID, enum.Rpc_friend_get_friend_list, func(buf *common.NetPack) {
	}, func(recvBuf *common.NetPack) {
		cnt := recvBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			friendId := recvBuf.ReadUInt32()

			//通告好友我下线了
			netConfig.CallRpcGateway(friendId, enum.Rpc_client_friend_logout, func(buf *common.NetPack) {
				buf.WriteUInt32(self.owner.AccountID)
			}, nil)
		}
	})
}

// ------------------------------------------------------------
// --
