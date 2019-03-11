package msg

type Retcode_ack struct {
	Retcode int    `json:"retcode"` //0代表成功，其他失败
	Msg     string `json:"msg"`     //响应信息（与retcode对应）
}
type Json_ack struct {
	Retcode_ack
	Data string `json:"data"`
}

// ------------------------------------------------------------
// 预下单回复【各渠道可能回复数据不同】
type Pre_buy_ack interface {
	SetRetcode(int)
	SetMsg(string)
	SetOrderId(string)
	HandleOrder(*TOrderInfo) bool
}
type Pre_buy struct { //默认
	Retcode_ack
	Order_id string `json:"order_id"`
}

func (self *Pre_buy) SetRetcode(v int)             { self.Retcode = v }
func (self *Pre_buy) SetMsg(v string)              { self.Msg = v }
func (self *Pre_buy) SetOrderId(v string)          { self.Order_id = v }
func (self *Pre_buy) HandleOrder(*TOrderInfo) bool { return true }

// ------------------------------------------------------------
// 订单查询回复
type Query_order_ack struct {
	Retcode int    `json:"retcode"` //0代表成功，其他失败
	Msg     string `json:"msg"`     //响应信息（与retcode对应）
	Order   struct {
		Pf_id       string `json:"pf_id"`       //平台名（如oppo）
		Pk_id       string `json:"pk_id"`       //包id
		Pay_id      int    `json:"pay_id"`      //支付商id（不同平台下的同一种支付渠道pay_id必须一样）
		Server_id   int    `json:"server_id"`   //服务器id（只有一个服务器的话，默认传1）
		Item_id     int    `json:"item_id"`     //物品id
		Item_name   string `json:"item_name"`   //物品名（中文需要urlencode）
		Item_count  int    `json:"item_count"`  //物品数量
		Item_price  int    `json:"item_price"`  //物品价格（单位是分）
		Total_price int    `json:"total_price"` //物品总价格（单位是分）
		Currency    string `json:"currency"`    //货币（在大陆，直接填RMB即可）
		Ip          string `json:"ip"`          //
		Status      int    `json:"status"`      //1成功 0失败
		Can_send    int    `json:"can_send"`    //1能发货
	} `json:"order"` //只有retcode是0才会有值，json
}
type Order_unfinished_ack struct {
	Orders []UnfinishedOrder `json:"orders"`
}
type UnfinishedOrder struct {
	Order_id       string `json:"order_id"`
	Third_order_id string `json:"third_order_id"`
	Item_id        int    `json:"item_id"`     //物品id
	Item_name      string `json:"item_name"`   //物品名（中文需要urlencode）
	Item_count     int    `json:"item_count"`  //物品数量
	Item_price     int    `json:"item_price"`  //物品价格（单位是分）
	Total_price    int    `json:"total_price"` //物品总价格（单位是分）
	Currency       string `json:"currency"`    //货币（在大陆，直接填RMB即可）
	Status         int    `json:"status"`      //1成功 0失败
	Can_send       int    `json:"can_send"`    //1能发货
}
