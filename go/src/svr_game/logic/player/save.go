/***********************************************************************
* @ Http玩家通用存档数据(依赖Http的同步阻塞特性)
* @ brief
	1、同client的SaveData.cs配合，实现自动云存档
	2、金币、钻石...敏感数据加密，防止用户窜改
	3、密钥由后台生成，后台每次从数据库载入存档数据时，重新生成密钥
	4、加密、解密由客户端自己负责，后台不知道具体数据格式，无法针敏感数据解密
	5、客户端上传的存档数据，应是解密后的，保障明文入库；网络层有自己的加密算法，用户不好窜改

* @ 存档同步
	1、存档模块带有自增ID，client每次修改存档触发，并上传后台
	2、若本地ID、服务器ID不一致，允许用户选择？还是做成游戏设置，默认本地覆盖远端？
	3、什么时候同步存档呢？修改即同步，网络不好可能引起卡顿。最好在回到主界面之类地方时整块同步

* @ author zhoumf
* @ date 2017-9-14
***********************************************************************/
package player

import (
	"common"
	"dbmgo"
	"generate_out/rpc/enum"
	"netConfig"
	"svr_game/api"
)

type TSaveClient struct {
	PlayerID uint32 `bson:"_id"`
	MAC      string //CloudArchive对应的机器码
}

// -------------------------------------
// -- 框架接口
func (self *TSaveClient) InitAndInsert(player *TPlayer) {
	self.PlayerID = player.PlayerID
	dbmgo.InsertToDB("Save", self)
}
func (self *TSaveClient) LoadFromDB(player *TPlayer) {
	if !dbmgo.Find("Save", "_id", player.PlayerID, self) {
		self.InitAndInsert(player)
	}
}
func (self *TSaveClient) WriteToDB() { dbmgo.UpdateSync("Save", self.PlayerID, self) }
func (self *TSaveClient) OnLogin() {
}
func (self *TSaveClient) OnLogout() {
}

// -------------------------------------
// -- Rpc
func Rpc_game_bind_cloud_archive(req, ack *common.NetPack) {
	account := req.ReadString()
	passwd := req.ReadString()
	force := req.ReadInt8()
	mac := req.ReadString()

	player := _Handle_Account_To_Player(account, passwd)
	if player == nil || (player.Save.MAC != "" && force == 0) {
		return
	}
	player.Save.MAC = mac
	player.Save.WriteToDB()
}
func Rpc_game_get_cloud_archive(req, ack *common.NetPack) {
	account := req.ReadString()
	passwd := req.ReadString()

	player := _Handle_Account_To_Player(account, passwd)
	if player == nil {
		ack.WriteString("")
		return
	}
	ack.WriteString(player.Save.MAC)
}

// -------------------------------------
// -- 辅助函数
func _Handle_Account_To_Player(account, passwd string) (ret *TPlayer) { //Notice:(┯_┯)函数依赖http同步阻塞特性
	accountId := uint32(0)
	api.CallRpcLogin(enum.Rpc_login_account_login, func(buf *common.NetPack) {
		buf.WriteInt(netConfig.G_Local_SvrID)
		buf.WriteString(account)
		buf.WriteString(passwd)
	}, func(recvBuf *common.NetPack) {
		err := recvBuf.ReadInt8()
		if err > 0 {
			accountId = recvBuf.ReadUInt32()
		}
	})
	if accountId == 0 {
		return
	}
	ret = FindWithDB_AccountId(accountId)
	if ret == nil {
		ret = AddNewPlayer(accountId, "")
	}
	return
}
