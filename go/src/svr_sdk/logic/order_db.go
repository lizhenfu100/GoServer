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
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
)

type TOrderInfo struct {
	Order_id      string `bson:"_id"`
	Pf_id         string //平台名（如oppo）
	Pk_id         string //包id（限定2位）
	Pay_id        int    //支付商id（不同平台下的同一种支付渠道pay_id必须一样）
	Op_id         int    //运营商 注：1代表移动。2代表联通 3代表电信
	App_id        string //应用id
	Server_id     int    //服务器id（只有一个服务器的话，默认传1）
	Account       string //账号（没有必须传空字符）
	Third_account string //第三方账号，相当于登录账号
	Role_id       string //角色id
	Item_id       int    //物品id
	Item_name     string //物品名（中文需要urlencode）
	Item_count    int    //物品数量
	Item_price    int    //物品价格（单位是分）
	Total_price   int    //物品总价格（单位是分）
	Currency      string //货币（在大陆，直接填RMB即可）
	Code          string //计费点
	Imsi          string
	Imei          string
	Ip            string
	Net           string //网络类型 CMNET， WIFI等
	Status        int    //1成功 0失败（第三方通告）
	Can_send      int    //1能发货 （client发货后置0）
}

var (
	g_order_map = make(map[string]*TOrderInfo, 1024)
)

func CacheOrder(ptr *TOrderInfo) {
	g_order_map[ptr.Order_id] = ptr
}
func FindOrder(orderId string) *TOrderInfo {
	if ptr, ok := g_order_map[orderId]; ok {
		return ptr
	}
	return nil
}
func DB_Save_Order(orderId string) bool {
	if pInfo, ok := g_order_map[orderId]; ok {
		if dbmgo.InsertSync("Order", pInfo) { //防止重复订单
			return true
		}
	}
	return false
}
func InitDB() {
	//载入所有未完成订单
	var list []TOrderInfo
	dbmgo.FindAll("Order", bson.M{
		"status":   bson.M{"$eq": 1},
		"can_send": bson.M{"$ne": 0},
	}, &list)
	for i := 0; i < len(list); i++ {
		CacheOrder(&list[i])
	}
}
