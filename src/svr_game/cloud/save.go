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
package cloud

import (
	"common"
	"common/math"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
)

type TSaveClient struct {
	MAC         string `bson:"_id"`
	ChangeTimes int32  //Json数据修改次数
	PublicKey   int32  //敏感数据加密公钥
	JsonData    string //无加密，明文，方便数据库查看
	// KV          map[string]interface{}
}

// -------------------------------------
// -- Rpc
func Rpc_game_get_cloud_archive_change_times(req, ack *common.NetPack) {
	mac := req.ReadString()

	self, _ := LoadSaveDataFromDB(mac)

	ack.WriteInt32(self.ChangeTimes)
}
func Rpc_game_upload_save_data(req, ack *common.NetPack) {
	mac := req.ReadString()
	times := req.ReadInt32()
	key := req.ReadInt32()
	data := req.ReadString()

	self, ok := LoadSaveDataFromDB(mac)

	if self.PublicKey != key || mac == "" {
		ack.WriteInt32(-1)
		return
	}

	self.MAC = mac
	self.ChangeTimes = times
	self.JsonData = data
	self.PublicKey = int32(math.RandBetween(0xff, 0xfffffff))

	ack.WriteInt32(self.PublicKey) //新公钥

	if ok {
		dbmgo.UpdateToDB("CloudArchive", bson.M{"_id": self.MAC}, bson.M{"$set": bson.M{
			"changetimes": self.ChangeTimes,
			"publickey":   self.PublicKey,
			"jsondata":    self.JsonData}})
	} else {
		dbmgo.InsertToDB("CloudArchive", self)
	}
}
func Rpc_game_download_save_data(req, ack *common.NetPack) {
	mac := req.ReadString()

	self, ok := LoadSaveDataFromDB(mac)

	if !ok {
		ack.WriteInt32(-1)
		return
	}

	self.PublicKey = int32(math.RandBetween(0xff, 0xfffffff))
	dbmgo.UpdateToDB("CloudArchive", bson.M{"_id": self.MAC}, bson.M{"$set": bson.M{"publickey": self.PublicKey}})

	ack.WriteInt32(self.PublicKey) //新公钥
	ack.WriteInt32(self.ChangeTimes)
	ack.WriteString(self.JsonData)
}

// -------------------------------------
// -- 辅助函数
func LoadSaveDataFromDB(mac string) (*TSaveClient, bool) {
	data := new(TSaveClient)
	ok := dbmgo.Find("CloudArchive", "_id", mac, data)
	return data, ok
}
