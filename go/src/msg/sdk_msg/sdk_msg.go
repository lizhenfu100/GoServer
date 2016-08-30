package sdk_msg

type Msg_create_recharge_order_Req struct { // 消息：/create_recharge_order
	SessionKey string

	PlayerID     int32
	OrderID      string
	Channel      string //渠道名
	PlatformEnum byte   //Android、IOS
	ChargeCsvID  int    //充值表ID
}
type Msg_create_recharge_order_Ack struct {
	RetCode int
}
type Msg_recharge_success struct { // 消息：/sdk_recharge_success
	PlayerID    int32
	ChargeCsvID int //充值表ID
	RMB         int //第三方发来的充值数
}

type SDKMsg_create_recharge_order_Req struct { // 消息：/create_recharge_order
	GamesvrID    int
	PlayerID     int32
	OrderID      string
	Channel      string //渠道名
	PlatformEnum byte   //Android、IOS
	ChargeCsvID  int    //充值表ID
}
type SDKMsg_create_recharge_order_Ack struct {
	RetCode int
}

type SDKMsg_recharge_result struct { // 消息：/sdk_recharge_success
	OrderID      string
	ThirdOrderID string
	Channel      string //渠道名
	PlayerID     int32
	RMB          int //充了多少钱
}
