package conf

//服务器配置数据
var (
	Version        string
	VerifyLoginUrl string //验证登录的URL

	//账号服
	AccountDbName string
	AccountDbAddr string

	//游戏服
	GameDbName string = "local"
	GameDbAddr string = "127.0.0.1:27017"

	//日志服
	LogSvrLogLevel int
	LogSvrFlushCnt int
	LogFileType    int    //日志文件类型
	LogFileName    string //日志文件名
)
