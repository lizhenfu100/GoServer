package msg

import (
	"common"
	"dbmgo"
	"fmt"

	"gopkg.in/mgo.v2/bson"
	"svr_game/logic/player"
)

type TBlackMarketGoods struct {
	ID    int
	IsBuy bool
}
type TBlackMarketModule struct {
	PlayerID int32 `bson:"_id"`

	Goods       []TBlackMarketGoods //! 商品
	RefreshTime int64               //! 刷新时间

	IsOpen   bool   //! 是否开启
	ResetDay uint32 //! 隔天刷新
}

// var g_test_mongodb TBlackMarketModule

func Rpc_test_mongodb(req, ack *common.ByteBuffer) interface{} {
	switch req.ReadByte() {
	case 1:
		{
			fmt.Println("CreateData")
			// g_test_mongodb.CreateData()
			ptr := player.AddNewPlayer(233, "zhoumf")
			fmt.Println(ptr)
			mail1 := ptr.Mail.CreateMail("title", "from", "content", common.IntPair{11, 2})
			ptr.Mail.CreateMail("title2", "from2", "content2", common.IntPair{22, 3})
			ptr.Mail.DelMail(mail1.ID)
		}
	case 2:
		{
			fmt.Println("UpdateData")
			// g_test_mongodb.UpdateData()
		}
	case 3:
		{
			fmt.Println("DelData")
			// g_test_mongodb.DelData()
		}
	default:
		{
			FindData()
		}
	}
	return nil
}
func (self *TBlackMarketModule) CreateData() {
	self.PlayerID = 233
	self.RefreshTime = 1234567
	self.ResetDay = 100
	self.IsOpen = true
	self.Goods = append(self.Goods, TBlackMarketGoods{7, false})
	dbmgo.InsertToDB("PlayerBlackMarket", self)
}
func FindData() {
	ptr := &TBlackMarketModule{}
	dbmgo.Find("PlayerBlackMarket", "_id", 233, ptr)
	fmt.Println(*ptr)
}
func (self *TBlackMarketModule) UpdateData() {
	self.ResetDay = 777
	self.Goods = append(self.Goods, TBlackMarketGoods{4, true})
	dbmgo.UpdateToDB("PlayerBlackMarket", bson.M{"_id": self.PlayerID}, bson.M{"$set": bson.M{
		"goods":    self.Goods,
		"resetday": self.ResetDay}})
}
func (self *TBlackMarketModule) DelData() {
	go dbmgo.RemoveSync("PlayerBlackMarket", bson.M{"_id": self.PlayerID})
}
