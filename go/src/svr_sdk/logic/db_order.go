/***********************************************************************
* @ 充值订单数据库
* @ brief
    1、gamesvr先通知SDK进程，建立新充值订单

    2、第三方充值信息到达后，验证是否为有效订单，通过后入库

* @ author zhoumf
* @ date 2016-8-18
***********************************************************************/
package logic

import (
	// "mongodb"
	"time"
	// "gopkg.in/mgo.v2/bson"
)

type TRechargeOrder struct {
	OrderID      string `bson:"_id"`
	ThirdOrderID string
	GamesvrID    int
	PlayerID     int32
	AccountID    int32
	CreateTime   string
	FinishTime   string
	Status       bool
	Channel      string //渠道名
	PlatformEnum byte   //Android、IOS
	RMB          int
	Content      string //JSON数据

	chargeCsvID int
}

var (
	g_order_map = make(map[string]*TRechargeOrder, 1024)
)

func CacheRechargeOrder(pOrder *TRechargeOrder) {
	pOrder.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	g_order_map[pOrder.OrderID] = pOrder
}
func DB_Save_RechargeOrder(orderID, thirdOrderID, content string, rmb int) *TRechargeOrder {
	if pInfo, ok := g_order_map[orderID]; ok {
		pInfo.RMB = rmb
		pInfo.Content = content
		pInfo.ThirdOrderID = thirdOrderID
		pInfo.FinishTime = time.Now().Format("2006-01-02 15:04:05")

		// if mongodb.InsertToDB(appconfig.GameDbName, "RechargeOrder", pInfo) { //防止重复订单
		delete(g_order_map, orderID)
		return pInfo
		// }
	}
	return nil
}
