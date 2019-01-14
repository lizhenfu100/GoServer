package conf

const (
	FPS_GameSvr      = 1000 / 20 //服务器帧率
	HandPick_GameSvr = false     //玩家手选区服

	// 特殊标记
	Flag_Client_ReLogin = 0xFFFFFFFF
	Flag_Compress       = 0x80000000

	// 通信的子功能开关
	Is_Msg_Compress     = false //消息压缩
	Is_Msg_Encode       = false //消息加密
	Open_Http_To_Client = false //Http推送

	// 统计信息
	Open_Calc_QPS = false //后台qps
)
