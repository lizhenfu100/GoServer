/***********************************************************************
* @ 多进程服务器架构
* @ brief
	1、主逻辑游戏服使用http同Client通信

	2、服务器进程间用tcp通信

* @ framework
	1、游戏服，一个玩家区服一个GameSvr
	2、战斗服，可动态扩展。可多个战斗服对应一个游戏服
	3、唯一的支付SDK、唯一的账号服
	4、唯一的Center，其它进程需在Center注册，管理后台所有进程

* @ reboot
	1、【1-1】关系中的"client"重启：game每次均会连接battle
	2、【1-1】关系中的"server"重启：battle(tcp)重启，game的client.ConnectToSvr能检查到失败，循环重连
	3、【1-N】关系中的"N"重启：game每次均会去sdk注册
	4、【1-N】关系中的"1"重启：http_server.go会本地存储注册地址，重启时载入

* @ author zhoumf
* @ date 2016-8-11
***********************************************************************/
package netConfig

import (
	"common"
	"fmt"
	"http"
	"strconv"
	"tcp"
)

// Notice：临时新增battle
// 1、先增加battle的配置（见注释）
// 2、执行bin/temp_svr.bat，在命令行输入"addsvrto 3"，通知game(http)来连接
type TNetConfig struct {
	Module     string
	SvrID      int
	IP         string // 内部局域网IP
	OutIP      string
	TcpPort    int
	HttpPort   int
	Maxconn    int
	ConnectLst []string // 待连接的模块名
}

var G_SvrNetCfg []TNetConfig = nil //见配表conf_net.csv

func GetNetCfg(module string, pSvrID *int) *TNetConfig { //负ID表示自动找首个
	for i := 0; i < len(G_SvrNetCfg); i++ {
		cfg := &G_SvrNetCfg[i]
		if cfg.Module == module && (*pSvrID < 0 || cfg.SvrID == *pSvrID) {
			*pSvrID = cfg.SvrID
			return cfg
		}
	}
	return nil
}
func GetLocalNetCfg() *TNetConfig {
	return GetNetCfg(G_Local_Module, &G_Local_SvrID)
}
func GetAddr(module string, svrID int) string {
	if cfg := GetNetCfg(module, &svrID); cfg != nil {
		if cfg.HttpPort > 0 {
			return fmt.Sprintf("http://%s:%d", cfg.IP, cfg.HttpPort)
		} else if cfg.TcpPort > 0 {
			return fmt.Sprintf("%s:%d", cfg.IP, cfg.TcpPort)
		} else {
			return ""
		}
	} else {
		return ""
	}
}

var (
	G_Cfg_Remote_TcpConn = make(map[common.KeyPair]*tcp.TCPClient) //本模块，对其它模块的tcp连接
	G_Local_Module       string
	G_Local_SvrID        int
)

func CreateNetSvr(module string, svrID int) bool {
	//1、找到当前的配置信息
	selfCfg := GetNetCfg(module, &svrID)
	if selfCfg == nil {
		print(fmt.Sprintf("%s-%d: have none SvrNetCfg!!!\n", module, svrID))
		return false
	}

	G_Local_Module = module
	G_Local_SvrID = svrID

	//2、连接/注册其它模块
	for _, destModule := range selfCfg.ConnectLst {
		for i := 0; i < len(G_SvrNetCfg); i++ {
			destCfg := &G_SvrNetCfg[i]
			if destCfg.Module == destModule {
				if destCfg.HttpPort > 0 {
					http.RegistToSvr(
						fmt.Sprintf("http://%s:%d/", destCfg.IP, destCfg.HttpPort),
						fmt.Sprintf("http://%s:%d/", selfCfg.IP, selfCfg.HttpPort),
						module,
						selfCfg.SvrID)
				} else if destCfg.TcpPort > 0 {
					client := new(tcp.TCPClient)
					client.ConnectToSvr(
						fmt.Sprintf("%s:%d", destCfg.IP, destCfg.TcpPort),
						module,
						selfCfg.SvrID)
					//Notice：client.ConnectToSvr是异步过程，这里返回的client.TcpConn还是空指针，不能保存*tcp.TCPConn
					G_Cfg_Remote_TcpConn[common.KeyPair{destCfg.Module, destCfg.SvrID}] = client
				} else {
					fmt.Println(destCfg.Module + ": have none HttpPort|TcpPort!!!")
				}
			}
		}
	}

	//3、开启本模块网络服务(Busy Loop)
	if selfCfg.HttpPort > 0 {
		http.NewHttpServer(":" + strconv.Itoa(selfCfg.HttpPort))
	} else if selfCfg.TcpPort > 0 {
		tcp.NewTcpServer(":"+strconv.Itoa(selfCfg.TcpPort), selfCfg.Maxconn)
	} else {
		fmt.Println(module + ": have none HttpPort|TcpPort!!!")
	}
	return true
}

func GetHttpAddr(destModule string, destSvrID int) string { //Notice：应用层cache住结果，避免每次都查找
	if destCfg := GetNetCfg(destModule, &destSvrID); destCfg != nil {
		selfCfg := GetLocalNetCfg()

		for _, v := range selfCfg.ConnectLst {
			if v == destModule && destCfg.HttpPort > 0 {
				// game(n) - sdk(1)
				return fmt.Sprintf("http://%s:%d/", destCfg.IP, destCfg.HttpPort)
			}
		}
	} else {
		fmt.Println(destModule + ": have none SvrNetCfg!!!")
		return ""
	}

	// sdk(1) - game(n)
	return http.FindRegModuleAddr(destModule, destSvrID)
}
func GetTcpConn(destModule string, destSvrID int) *tcp.TCPConn { //Notice：应用层cache住结果，避免每次都查找
	if destCfg := GetNetCfg(destModule, &destSvrID); destCfg != nil {
		selfCfg := GetLocalNetCfg()

		for _, v := range selfCfg.ConnectLst {
			if v == destModule {
				// game(c) - battle(s)
				return G_Cfg_Remote_TcpConn[common.KeyPair{destModule, destSvrID}].TcpConn
			}
		}
	} else {
		fmt.Println(destModule + ": have none SvrNetCfg!!!")
		return nil
	}

	// battle(s) - game(c)
	return tcp.FindRegModuleConn(destModule, destSvrID)
}

// 已验证：此函数失败，resp是nil，那resp.Body.Close()就不能无脑调了
// resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
// resp.Body.Close()
