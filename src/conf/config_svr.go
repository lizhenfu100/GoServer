package conf

const (
	GameName         = "soulnet"
	FPS_GameSvr      = 1000 / 20 //游戏服帧率
	FPS_OtherSvr     = 1000 / 10 //其它服帧率
	HandPick_GameSvr = false     //玩家手选区服
	HaveClientSave   = false     //是否启用云存档
	IsTcpGame        = true      //game节点是tcp还是http

	// 特殊标记
	Flag_Client_ReLogin = 0xFFFFFFFF
	Flag_Compress       = 0x80000000

	// 通信的子功能开关
	Is_Msg_Compress   = false //消息压缩
	Is_Msg_Encode     = false //消息加密
	Is_Http_To_Client = true  //Http推送

	// 测试标记
	TestFlag_CalcQPS = false //后台qps

	// GM相关
	GM_Passwd = "chillyroom_gm_*"
)

var SvrCsv struct {
	// 数据库
	DBuser   string
	DBpasswd string

	// 邮箱
	EmailUser     []string
	EmailPasswd   []string
	EmailHost     string
	EmailPort     int
	EmailLanguage string //默认语言，参见language.go
	// 微信
	WechatCorpId  string //企业id
	WechatSecret  string //应用的Secret
	WechatTouser  string //接收者,多个用‘|’分隔
	WechatAgentId int    //应用id
}
