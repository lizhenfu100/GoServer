package test

import (
	"common"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
	_ "svr_client/test/init"
	"tcp"
	"testing"
)

type loginData struct {
	account    string
	passwd     string
	accountId  uint32
	ip         string
	port       uint16
	token      uint32
	playerName string
	playerId   uint32
}

var (
	g_gameName  = "HappyDiner"
	g_loginData = loginData{
		account:    "166@qq.com",
		passwd:     "123123",
		playerName: "test",
	}
	g_centerAddr = netConfig.GetHttpAddr("center", 1)
	g_loginList  = []string{
		"http://192.168.1.111:7030",
		"http://52.14.1.205:7030",    //1 北美
		"http://13.229.215.168:7030", //2 亚洲
		"http://54.94.211.178:7030",  //3 南美
		"http://18.185.80.202:7030",  //4 欧洲
		"http://39.96.196.250:7030",  //5 华北
		"http://47.106.35.74:7030",   //6 华南
	}
	g_loginAddr = g_loginList[0]
	g_version = ""
)

// go test -v ./src/svr_client/test/login_test.go
func Test_login_init(t *testing.T) {
	//for i := 0; i < 100; i++ {
	//	g_loginData.account = fmt.Sprintf("tester%dchillyroom.com", i)
	//	g_loginData.passwd = fmt.Sprintf("tester%d", i)
	//	Test_account_reg(t)
	//	Test_account_login(t)
	//}
}

// ------------------------------------------------------------
func Test_get_gamesvr_list(t *testing.T) {
	http.CallRpc(g_loginAddr, enum.Rpc_login_get_game_list, func(buf *common.NetPack) {
		buf.WriteString(g_version)
	}, func(backBuf *common.NetPack) {
		cnt := backBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			id := backBuf.ReadInt()
			svrName := backBuf.ReadString()
			outIp := backBuf.ReadString()
			port := backBuf.ReadUInt16()

			onlineCnt := backBuf.ReadInt32()
			fmt.Println("-------GamesvrList:", id, svrName, outIp, port, onlineCnt)
		}
	})
}
func tTest_player_gameinfo_addr(t *testing.T) {
	http.CallRpc(g_centerAddr, enum.Rpc_center_player_login_addr, func(buf *common.NetPack) {
		buf.WriteString(g_gameName)
		buf.WriteString(g_loginData.account)
	}, func(backBuf *common.NetPack) {
		if e := backBuf.ReadUInt16(); e == err.Success {
			loginIp := backBuf.ReadString()
			loginPort := backBuf.ReadUInt16()
			gameIp := backBuf.ReadString()
			gamePort := backBuf.ReadUInt16()
			fmt.Println("GameInfo Addr: ", loginIp, loginPort, gameIp, gamePort)
		} else {
			fmt.Println("ameInfo Addr Err: ", e)
		}
	})
}

// ------------------------------------------------------------
// -- 注册
func tTest_account_reg(t *testing.T) {
	http.CallRpc(g_loginAddr, enum.Rpc_login_relay_to_center, func(buf *common.NetPack) {
		buf.WriteUInt16(enum.Rpc_center_account_reg)
		buf.WriteString(g_loginData.account)
		buf.WriteString(g_loginData.passwd)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("-------AccountReg errCode:", errCode)
	})
}

// ------------------------------------------------------------
// -- 登录，三种方式
func tTest_account_login(t *testing.T) {
	http.CallRpc(g_loginAddr, enum.Rpc_login_account_login, func(buf *common.NetPack) {
		buf.WriteString(g_version)
		buf.WriteString(g_gameName)
		//buf.WriteInt32(gameSvrId); //玩家手选区服方式
		buf.WriteString(g_loginData.account)
		buf.WriteString(g_loginData.passwd)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		if errCode == err.Success {
			g_loginData.accountId = backBuf.ReadUInt32()
			g_loginData.ip = backBuf.ReadString()
			g_loginData.port = backBuf.ReadUInt16()
			g_loginData.token = backBuf.ReadUInt32()
			fmt.Println("-------AccountLogin ok:", "继续登录游戏服...有三种情况：")
			fmt.Println("    一、直连HttpGamesvr\n    二、直连TcpGamesvr\n    三、Gateway接管")
			g_loginData.LoginGamesvr_http()
			//g_loginData.LoginGamesvr_tcp()
			//g_loginData.LoginGateway()
		} else {
			fmt.Println("-------AccountLogin errCode:", errCode)
		}
	})
}
func (self *loginData) LoginGamesvr_http() { //方式一：直接登录Http游戏服
	var playerRpc *http.PlayerRpc
	gameAddr := http.Addr(self.ip, self.port)
	http.CallRpc(gameAddr, enum.Rpc_game_login, func(buf *common.NetPack) {
		buf.WriteUInt32(self.accountId)
		buf.WriteUInt32(self.token)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("-------GameLogin errCode:", errCode)
		if errCode == err.Success {
			self.playerId = backBuf.ReadUInt32()
			self.playerName = backBuf.ReadString()
			fmt.Println("-------GameLogin ok:", self.accountId, self.playerId, self.playerName)
			playerRpc = http.NewPlayerRpc(gameAddr, self.accountId)
		} else if errCode == err.Account_have_none_player {
			fmt.Println("-------GameLogin:", "请创建角色")
			http.CallRpc(gameAddr, enum.Rpc_game_create_player, func(buf *common.NetPack) {
				buf.WriteUInt32(self.accountId)
				buf.WriteString(self.playerName)
			}, func(backBuf *common.NetPack) {
				if playerId := backBuf.ReadUInt32(); playerId > 0 {
					self.playerId = playerId
					fmt.Println("-------CreatePlayer ok:", playerId, self.playerName)
					playerRpc = http.NewPlayerRpc(gameAddr, self.accountId)
				} else {
					fmt.Println("-------CreatePlayer:", "创建角色失败")
				}
			})
		}
	})
}
func (self *loginData) LoginGamesvr_tcp() { //方式二：直接登录Tcp游戏服
	gamesvr := new(tcp.TCPClient)
	gamesvr.OnConnect = func(conn *tcp.TCPConn) {
		conn.CallRpc(enum.Rpc_game_login, func(buf *common.NetPack) {
			buf.WriteUInt32(self.accountId)
			buf.WriteUInt32(self.token)
		}, func(backBuf *common.NetPack) {
			errCode := backBuf.ReadUInt16()
			fmt.Println("-------GameLogin errCode:", errCode)
			if errCode == err.Success {
				self.playerId = backBuf.ReadUInt32()
				self.playerName = backBuf.ReadString()
				fmt.Println("-------GameLogin ok:", self.playerId, self.playerName)
			} else if errCode == err.Account_have_none_player {
				fmt.Println("-------GameLogin:", "请创建角色")
				name := "test_zhoumf233"
				conn.CallRpc(enum.Rpc_game_create_player, func(buf *common.NetPack) {
					buf.WriteUInt32(self.accountId)
					buf.WriteString(name)
				}, func(backBuf *common.NetPack) {
					if playerId := backBuf.ReadUInt32(); playerId > 0 {
						self.playerId = playerId
						self.playerName = name
						fmt.Println("-------CreatePlayer ok:", playerId, name)
					} else {
						fmt.Println("-------CreatePlayer:", "创建角色失败")
					}
				})
			}
		})
	}
	gamesvr.ConnectToSvr(tcp.Addr(self.ip, self.port), meta.G_Local)
}
func (self *loginData) LoginGateway() { //方式三：Gateway网关转接消息
	gateway := new(tcp.TCPClient)
	gateway.OnConnect = func(conn *tcp.TCPConn) {
		conn.CallRpc(enum.Rpc_gateway_login, func(buf *common.NetPack) {
			buf.WriteUInt32(self.accountId)
			buf.WriteUInt32(self.token)
		}, func(backBuf *common.NetPack) {
			errCode := backBuf.ReadUInt16()
			fmt.Println("-------GatewayLogin errCode:", errCode)
			if errCode == err.Success {
				conn.CallRpc(enum.Rpc_gateway_relay_game_login, func(buf *common.NetPack) {}, func(backBuf *common.NetPack) {
					errCode := backBuf.ReadUInt16()
					fmt.Println("-------GameLogin errCode:", errCode)
					if errCode == err.Success {
						self.playerId = backBuf.ReadUInt32()
						self.playerName = backBuf.ReadString()
						fmt.Println("-------GameLogin ok:", self.playerId, self.playerName)
					} else if errCode == err.Account_have_none_player {
						fmt.Println("-------GameLogin:", "请创建角色")
						name := "test_zhoumf233"
						conn.CallRpc(enum.Rpc_gateway_relay_game_create_player, func(buf *common.NetPack) {
							buf.WriteString(name)
						}, func(backBuf *common.NetPack) {
							if playerId := backBuf.ReadUInt32(); playerId > 0 {
								self.playerId = playerId
								self.playerName = name
								fmt.Println("-------CreatePlayer ok:", playerId, name)
							} else {
								fmt.Println("-------CreatePlayer:", "创建角色失败")
							}
						})
					}
				})
			}
		})
	}
	gateway.ConnectToSvr(tcp.Addr(self.ip, self.port), meta.G_Local)
}

// ------------------------------------------------------------
// --
