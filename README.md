# 多进程Go游戏服务器
##网络拓扑关系纯配置控制

核心的tcp、http是项目正在用的，效果还可以，虽然还没被外网玩家操过~

在此之上封了一层netConfig，便于网络连接关系的管理

各模块进程支持崩溃重启

##简介

1、核心的tcp、http是项目正在用的，效果还可以，虽然还没被外网玩家操过

2、在此之上封了一层netConfig，便于网络连接关系的管理

3、支持崩溃重启

	(1)【1-1】关系中的"client"重启：game每次均会连接battle

	(2)【1-1】关系中的"server"重启：battle(tcp)重启，game的client.ConnectToSvr能检查到失败，循环重连

	(3)【1-N】关系中的"N"重启：     game每次均会去sdk注册

	(4)【1-N】关系中的"1"重启：     "http_server.go"会本地存储注册地址，重启时载入

4、使用方便

[config_net.go](https://github.com/3workman/Sundry/tree/master/go/src/netConfig/config_net.go)
--------------

	(1)"config_net.go"中非常容易指定连接方式
```go
var G_SvrNetCfg = map[string]TSvrNetCfg{
	"sdk": {
		TAddrInfo{
			IP:       "127.0.0.1",
			OutIP:    "192.168.1.177",
			HttpPort: 7002,
		},
		[]string{},
	},
	"game": {
		TAddrInfo{
			IP:       "127.0.0.1",
			OutIP:    "192.168.1.177",
			HttpPort: 7010,
			SvrID:    1,
		},
		[]string{"sdk", "battle"}, // 连接sdk、battle
	},
	"battle": {
		TAddrInfo{
			IP:      "127.0.0.1",
			OutIP:   "192.168.1.177",
			TcpPort: 7030,
			Maxconn: 5000,
			SvrID:   1,
		},
		[]string{},
	},
	"client": {
		TAddrInfo{},
		[]string{"game", "sdk", "battle"}, // 连接game、sdk、battle
	},
}
```
	
	(2)统一了连接获取方式，业务层使用，只需加几个Cache接口即可
```go
var (
	g_cache_battle_conn *tcp.TCPConn
)
func SendToBattle(msgID uint16, msgdata []byte) {
	if g_cache_battle_conn == nil {
		g_cache_battle_conn = netConfig.GetTcpConn("battle", 0)
	}
	g_cache_battle_conn.WriteMsg(msgID, msgdata)
}
```
	
	(3)构建一个新的服务进程，配置完毕后，只两行代码+SendToModule接口就够啦
```go
func main() {
	//注册所有tcp消息处理方法
	RegBattleTcpMsgHandler()

	gamelog.Warn("----Battle Server Start-----")
	if netConfig.CreateNetSvr("battle") == false {
		gamelog.Error("----Battle NetSvr Failed-----")
	}
}
```

5、目前只配了game、battle、sdk、client网络模块，编译后运行bin目录的start_svr.bat可验证测试