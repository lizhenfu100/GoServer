/***********************************************************************
* @ 充值订单数据库
* @ brief
    1、gamesvr先通知SDK进程，建立新充值订单

    2、第三方充值信息到达后，验证是否为有效订单，通过后入库

* @ author zhoumf
* @ date 2016-8-18
***********************************************************************/
package msg

import (
	"common/timer"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

const KDBTable = "Order"

type TOrderInfo struct {
	Order_id       string `bson:"_id"`
	Third_order_id string
	Pf_id          string //平台名（如oppo）
	Pk_id          string //包id
	App_id         string //应用id
	Pay_id         int    //支付商id（不同平台下的同一种支付渠道pay_id必须一样）
	Op_id          int    //运营商 注：1代表移动 2代表联通 3代表电信
	Server_id      int    //服务器id（只有一个服务器的话，默认传1）
	Account        string //账号（没有必须传空字符）
	Third_account  string //第三方账号，相当于登录账号
	Role_id        string //角色id
	Item_name      string //物品名（中文需要urlencode）
	Item_id        int    //物品id
	Item_count     int    //物品数量
	Item_price     int    //物品价格（单位是分）
	Total_price    int    //物品总价格（单位是分）
	Currency       string //货币（在大陆，直接填RMB即可）
	Code           string //计费点
	Imsi           string
	Imei           string
	Ip             string
	Net            string //网络类型 CMNET， WIFI等
	Status         int    //1成功 0失败（第三方通告）
	Can_send       int    //1能发货 （client发货后置0）
	Time           int64
	Extra          string
}

var g_order_map sync.Map //<orderId, *TOrderInfo>

func CreateOrderInDB(ptr *TOrderInfo) bool {
	ptr.Order_id = fmt.Sprintf("%03d%s%06d", //生成订单号
		ptr.Pay_id,
		time.Now().Format("060102"),
		dbmgo.GetNextIncId("OrderId"))
	ptr.Time = time.Now().Unix()
	if dbmgo.InsertSync(KDBTable, ptr) {
		g_order_map.Store(ptr.Order_id, ptr)
		return true
	}
	return false
}
func FindOrder(orderId string) *TOrderInfo {
	if orderId == "" || orderId == "0" {

	} else if v, ok := g_order_map.Load(orderId); ok {
		return v.(*TOrderInfo)
	} else {
		ptr := new(TOrderInfo)
		if ok, _ := dbmgo.Find(KDBTable, "_id", orderId, ptr); ok {
			return ptr
		}
	}
	return nil
}
func ConfirmOrder(ptr *TOrderInfo) {
	ptr.Can_send = 0
	dbmgo.UpdateId(KDBTable, ptr.Order_id, bson.M{"$set": bson.M{"can_send": 0}})
	g_order_map.Delete(ptr.Order_id)
}
func OrderRange(f func(k, v interface{}) bool) { g_order_map.Range(f) }

func InitDB() {
	now := time.Now().Unix()
	//删除超过7天的无效订单
	dbmgo.RemoveAllSync(KDBTable, bson.M{
		"status": 0, "can_send": 0,
		"time": bson.M{"$lt": now - 7*24*3600},
	})
	//删除超过30天的
	dbmgo.RemoveAllSync(KDBTable, bson.M{"time": bson.M{"$lt": now - 30*24*3600}})
	//载入所有未完成的
	var list []TOrderInfo
	dbmgo.FindAll(KDBTable, bson.M{"can_send": 1}, &list)
	for i := 0; i < len(list); i++ {
		g_order_map.Store(list[i].Order_id, &list[i])
	}
	println("load active order form db: ", len(list))

	timer.G_TimerMgr.AddTimerSec(DeleteTimeOutOrder, 24*3600, 24*3600, -1)
}
func DeleteTimeOutOrder() { //删除内存中滞留一天的订单(完成的、无效的)
	timenow := time.Now().Unix()
	g_order_map.Range(func(k, v interface{}) bool {
		p := v.(*TOrderInfo)
		if p.Can_send == 0 && timenow-p.Time > 24*3600 {
			g_order_map.Delete(k)
		}
		return true
	})
}
