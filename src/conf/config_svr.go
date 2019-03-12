package conf

const (
	GameName         = "soulnet"
	FPS_GameSvr      = 1000 / 20 //游戏服帧率
	HandPick_GameSvr = false     //玩家手选区服
	HaveCllientSave  = false     //是否启用云存档

	// 特殊标记
	Flag_Client_ReLogin = 0xFFFFFFFF
	Flag_Compress       = 0x80000000

	// 通信的子功能开关
	Is_Msg_Compress     = false //消息压缩
	Is_Msg_Encode       = false //消息加密
	Open_Http_To_Client = false //Http推送

	// 测试标记
	TestFlag_CalcQPS = false //后台qps

	// GM相关
	GM_Passwd = "chillyroom_gm_*"
)

var SvrCsv struct {
	//数据库
	DBuser   string
	DBpasswd string

	// 邮箱服务
	//kUser, kPasswd = "515693380@qq.com", "afcoucpylyebbhjb"
	//kHost, kPort   = "smtp.qq.com", 465
	//kUser, kPasswd = "3workman@gmail.com", "zmf890104"
	//kHost, kPort   = "smtp.gmail.com", 465
	EmailUser   string
	EmailPasswd string
	EmailHost   string
	EmailPort   int

	WechatCorpId  string //企业id
	WechatSecret  string //应用的Secret
	WechatTouser  string //接收者,多个用‘|’分隔
	WechatAgentId int    //应用id
}
