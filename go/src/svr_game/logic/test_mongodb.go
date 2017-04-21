package logic

import (
	"common"
	"dbmgo"
	"fmt"
	"gamelog"
	"net/http"

	"gopkg.in/mgo.v2/bson"
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

var g_test_mongodb TBlackMarketModule

func Rpc_test_mongodb(w http.ResponseWriter, r *http.Request) {
	gamelog.Info("message: %s", r.URL.String())

	//! 接收信息
	msg := common.NewNetPackLen(int(r.ContentLength))
	r.Body.Read(msg.DataPtr)

	//! 创建回复
	defer func() {
		w.Write([]byte("ok"))
	}()

	switch msg.ReadByte() {
	case 1:
		{
			fmt.Println("CreateData")
			g_test_mongodb.CreateData()
		}
	case 2:
		{
			fmt.Println("UpdateData")
			g_test_mongodb.UpdateData()
		}
	case 3:
		{
			fmt.Println("DelData")
			g_test_mongodb.DelData()
		}
	default:
		{
			FindData()
		}
	}
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
