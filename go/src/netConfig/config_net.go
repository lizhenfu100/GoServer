/***********************************************************************
* @ 多进程服务器架构
* @ brief
	1、主逻辑游戏服使用http同Client通信

	2、服务器进程间用tcp通信

	3、未来扩展：Battle设计为多个，初始化完毕后http.Post自己的信息到Gamesvr（甚至能临时加机器）

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
	"fmt"
	"http"
	"strconv"
	"tcp"
)

type TAddrInfo struct {
	IP       string // 内部局域网IP
	OutIP    string
	TcpPort  int
	HttpPort int
	Maxconn  int
	SvrID    int
}
type TSvrNetCfg struct {
	Listen  []TAddrInfo
	Connect []string
}

// Notice：临时新增battle
// 1、先增加battle的配置（见注释），重新编译svr_battle.exe【TODO：做成配置文件就不用编译啦】
// 2、执行bin/temp_svr.bat，在命令行输入"addsvrto 3"，通知game(http)来连接
var G_SvrNetCfg = map[string]TSvrNetCfg{
	"account": {
		[]TAddrInfo{
			{
				IP:      "127.0.0.1",
				TcpPort: 7001,
				Maxconn: 5000,
			},
		},
		[]string{},
	},
	"sdk": {
		[]TAddrInfo{
			{
				IP:       "127.0.0.1",
				OutIP:    "192.168.1.177",
				HttpPort: 7002,
				SvrID:    1,
			},
		},
		[]string{},
	},
	"cross": {},

	"game": {
		[]TAddrInfo{
			{
				IP:       "127.0.0.1",
				OutIP:    "192.168.1.177",
				HttpPort: 7010,
				SvrID:    1,
			},
		},
		[]string{"sdk", "battle"}, // []string{"chat", "battle", "sdk"},
	},
	"chat": {
		[]TAddrInfo{
			{
				IP:      "127.0.0.1",
				OutIP:   "192.168.1.177",
				TcpPort: 7020,
				Maxconn: 5000,
			},
		},
		[]string{},
	},
	"battle": {
		[]TAddrInfo{
			{
				IP:      "127.0.0.1",
				OutIP:   "192.168.1.177",
				TcpPort: 7030,
				Maxconn: 5000,
				SvrID:   1,
			}, {
				IP:      "127.0.0.1",
				OutIP:   "192.168.1.177",
				TcpPort: 7031,
				Maxconn: 5000,
				SvrID:   2,
			},
			// { //临时增加的战斗服
			// 	IP:      "127.0.0.1",
			// 	OutIP:   "192.168.1.177",
			// 	TcpPort: 7032,
			// 	Maxconn: 5000,
			// 	SvrID:   3,
			// },
		},
		[]string{},
	},
	"client": {
		[]TAddrInfo{{}}, //配一个空的
		[]string{"game", "sdk", "battle"},
	},
}

func GetAddr(module string, svrID int) string {
	for _, v := range G_SvrNetCfg[module].Listen {
		if v.SvrID == svrID {
			if v.HttpPort > 0 {
				return fmt.Sprintf("http://%s:%d", v.IP, v.HttpPort)
			} else if v.TcpPort > 0 {
				return fmt.Sprintf("%s:%d", v.IP, v.TcpPort)
			} else {
				return ""
			}
		}
	}
	return ""
}

var (
	G_Cfg_Remote_TcpConn = make(map[tcp.TcpConnKey]*tcp.TCPClient) //本模块，对其它模块的tcp连接
	G_Local_Module       string
	G_Local_SvrID        int
)

func CreateNetSvr(module string, svrID int) bool {
	G_Local_Module = module
	G_Local_SvrID = svrID

	if cfg, ok := G_SvrNetCfg[module]; ok {

		//1、找到当前svrID的配置信息
		var selfCfg *TAddrInfo = nil
		for i := 0; i < len(cfg.Listen); i++ {
			if svrID == cfg.Listen[i].SvrID {
				selfCfg = &cfg.Listen[i]
				break
			}
		}
		if selfCfg == nil {
			print(fmt.Sprintf("%s-%d: have none SvrNetCfg!!!", module, svrID))
			return false
		}

		//2、连接/注册其它模块
		for _, v := range cfg.Connect {
			if cfg2, ok2 := G_SvrNetCfg[v]; ok2 {
				for _, destCfg := range cfg2.Listen {
					if destCfg.HttpPort > 0 {
						http.RegistToSvr(
							fmt.Sprintf("http://%s:%d", destCfg.IP, destCfg.HttpPort),
							fmt.Sprintf("http://%s:%d", selfCfg.IP, selfCfg.HttpPort),
							module,
							selfCfg.SvrID)
					} else if destCfg.TcpPort > 0 {
						client := &tcp.TCPClient{}
						client.ConnectToSvr(
							fmt.Sprintf("%s:%d", destCfg.IP, destCfg.TcpPort),
							module,
							selfCfg.SvrID)
						//Notice：client.ConnectToSvr是异步过程，这里返回的client.TcpConn还是空指针，不能保存*tcp.TCPConn
						G_Cfg_Remote_TcpConn[tcp.TcpConnKey{v, destCfg.SvrID}] = client
					} else {
						print(v + ": have none HttpPort|TcpPort!!!")
					}
				}
			} else {
				print(v + ": have none SvrNetCfg!!!")
				return false
			}
		}

		//3、开启本模块网络服务(Busy Loop)
		if selfCfg.HttpPort > 0 {
			http.NewHttpServer(":" + strconv.Itoa(selfCfg.HttpPort))
		} else if selfCfg.TcpPort > 0 {
			tcp.NewTcpServer(":"+strconv.Itoa(selfCfg.TcpPort), selfCfg.Maxconn)
		} else {
			print(module + ": have none HttpPort|TcpPort!!!")
		}
		return true
	}
	print(module + ": have none SvrNetCfg!!!")
	return false
}

func GetHttpAddr(destModule string, destSvrID int) string { //Notice：应用层cache住结果，避免每次都查找
	var destCfg *TAddrInfo = nil
	if cfg, ok := G_SvrNetCfg[destModule]; ok {
		if destSvrID <= 0 {
			destSvrID = cfg.Listen[0].SvrID
		}
		for i := 0; i < len(cfg.Listen); i++ {
			if destSvrID == cfg.Listen[i].SvrID {
				destCfg = &cfg.Listen[i]
				break
			}
		}
	} else {
		print(destModule + ": have none SvrNetCfg!!!")
		return ""
	}

	for _, v := range G_SvrNetCfg[G_Local_Module].Connect {
		if v == destModule && destCfg.HttpPort > 0 {
			// game(n) - sdk(1)
			return fmt.Sprintf("http://%s:%d", destCfg.IP, destCfg.HttpPort)
		}
	}

	// sdk(1) - game(n)
	return http.FindRegModuleAddr(destModule, destSvrID)
}
func GetTcpConn(destModule string, destSvrID int) *tcp.TCPConn { //Notice：应用层cache住结果，避免每次都查找
	if cfg, ok := G_SvrNetCfg[destModule]; ok {
		if destSvrID <= 0 {
			destSvrID = cfg.Listen[0].SvrID
		}
	} else {
		print(destModule + ": have none SvrNetCfg!!!")
		return nil
	}

	for _, v := range G_SvrNetCfg[G_Local_Module].Connect {
		if v == destModule {
			// game(c) - battle(s)
			return G_Cfg_Remote_TcpConn[tcp.TcpConnKey{destModule, destSvrID}].TcpConn
		}
	}

	// battle(s) - game(c)
	return tcp.FindRegModuleConn(destModule, destSvrID)
}

// 已验证：此函数失败，resp是nil，那resp.Body.Close()就不能无脑调了
// resp, err := http.Post(url, "text/HTML", bytes.NewReader(b))
// resp.Body.Close()
