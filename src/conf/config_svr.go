package conf

const (
	GameName       = "soulnet"
	FPS_GameSvr    = 1000 / 20 //游戏服帧率
	FPS_OtherSvr   = 1000 / 10 //其它服帧率
	Auto_GameSvr   = true      //自动选区服
	HaveClientSave = false     //是否启用云存档

	// 特殊标记
	Flag_Client_ReLogin = 0xFFFFFFFF
	Flag_Compress       = 0x80000000

	// 通信的子功能开关
	Is_Msg_Compress   = false //消息压缩
	Is_Msg_Encode     = false //消息加密
	Is_Http_To_Client = false //Http推送

	// 测试标记
	TestFlag_CalcQPS = false //后台qps

	// GM相关
	GM_Passwd = "chillyroom_gm_*"
)

//go:generate D:\server\bin\gen_conf.exe *svrCsv conf
type svrCsv struct {
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
	WechatAgentId int    //应用id
	// 短信
	SmsKeyId  string
	SmsSecret string
}

var (
	SvrList = []string{ //本项目包含哪类节点
		"shared_svr/svr_center",
		"shared_svr/svr_dns",
		"shared_svr/svr_file",
		"shared_svr/svr_friend",
		"shared_svr/svr_gateway",
		"shared_svr/svr_gm",
		"shared_svr/svr_login",
		"shared_svr/svr_sdk",
		"shared_svr/svr_relay",
		"shared_svr/svr_nric",
		"svr_cross",
		"svr_game",
	}
)
