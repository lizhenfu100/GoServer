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
	"common"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
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
	Game_id        string //游戏名称
	Account        string //账号（没有必须传空字符）
	Third_account  string //第三方账号，相当于登录账号
	Role_id        string //角色id
	Item_name      string //物品名
	Item_id        int    //物品id
	Item_count     int    //物品数量
	Item_price     int    //物品价格（单位是分）
	Total_price    int    //物品总价格（单位是分）
	Currency       string //货币类型
	Pay_channel    string //支付渠道，微信、支付宝...
	Code           string //计费点
	Imsi           string
	Imei           string //移动设备识别码
	Ip             string
	Net            string //网络类型 CMNET， WIFI等
	Status         int    //1成功 0失败（第三方通告）
	Can_send       int    //1能发货 （client发货后置0）
	Time           int64
	Mac            string //设备码
	Version_code   string //客户端版本号
	Extra          string
}

func CreateOrderInDB(ptr *TOrderInfo) error {
	if incId := dbmgo.GetNextIncId("OrderId"); incId > 0 {
		ptr.Order_id = fmt.Sprintf("%03d%s%06d", //生成订单号
			ptr.Pay_id,
			time.Now().Format("060102"),
			incId)
		ptr.Time = time.Now().Unix()
		return dbmgo.DB().C(KDBTable).Insert(ptr)
	}
	return common.Err("GetNextIncId")
}
func FindOrder(orderId string) *TOrderInfo {
	if orderId != "" && orderId != "0" {
		ptr := new(TOrderInfo)
		if ok, _ := dbmgo.Find(KDBTable, "_id", orderId, ptr); ok {
			return ptr
		}
	}
	return nil
}
func ConfirmOrder(ptr *TOrderInfo) { //确认发货
	ptr.Can_send = 0
	dbmgo.UpdateIdSync(KDBTable, ptr.Order_id, bson.M{"$set": bson.M{"can_send": 0}})
}

func ClearOldOrder() {
	now := time.Now().Unix()
	//删除超过7天的无效订单
	dbmgo.RemoveAllSync(KDBTable, bson.M{
		"status": 0, "can_send": 0,
		"time": bson.M{"$lt": now - 7*24*3600},
	})
	//删除超过30天的
	dbmgo.RemoveAllSync(KDBTable, bson.M{"time": bson.M{"$lt": now - 30*24*3600}})
}
