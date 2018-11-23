package conf

const (
	// 服务器帧率
	FPS_GameSvr = 1000 / 20

	// 特殊标记
	Flag_Client_ReLogin = 0xFFFFFFFF

	// 通信的子功能开关
	Is_Msg_Compress     = false //消息压缩
	Is_Msg_Encode       = false //消息加密
	Open_Http_To_Client = false //Http推送
)

var SvrCsv struct {
	// 数据库
	DBuser   string
	DBpasswd string
}
