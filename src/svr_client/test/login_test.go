package test

import (
	"common"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"tcp"
	"testing"
)

var (
	g_loginData = loginData{
		account: "zhoumf",
		passwd:  "123",
	}
	g_httpPlayer *http.PlayerRpc //http svr登录成功后，用以CallRpc业务功能
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

func Test_account_reg(t *testing.T) {
	loginAddr := netConfig.GetHttpAddr("login", 1)
	http.CallRpc(loginAddr, enum.Rpc_login_relay_to_center, func(buf *common.NetPack) {
		buf.WriteUInt16(enum.Rpc_center_account_reg)
		buf.WriteString(g_loginData.account)
		buf.WriteString(g_loginData.passwd)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("-------AccountReg errCode:", errCode)
	})
}
func Test_account_login(t *testing.T) {
	loginAddr := netConfig.GetHttpAddr("login", 1)
	http.CallRpc(loginAddr, enum.Rpc_login_account_login, func(buf *common.NetPack) {
		buf.WriteString("") //version
		buf.WriteString("") //游戏名称
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

// ------------------------------------------------------------
// -- 方式一：直接登录Http游戏服
func (self *loginData) LoginGamesvr_http() {
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
			fmt.Println("-------GameLogin ok:", self.playerId, self.playerName)
			g_httpPlayer = http.NewPlayerRpc(gameAddr, self.accountId)
		} else if errCode == err.Account_have_none_player {
			fmt.Println("-------GameLogin:", "请创建角色")
			name := "test_zhoumf233"
			http.CallRpc(gameAddr, enum.Rpc_game_create_player, func(buf *common.NetPack) {
				buf.WriteUInt32(self.accountId)
				buf.WriteString(name)
			}, func(backBuf *common.NetPack) {
				if playerId := backBuf.ReadUInt32(); playerId > 0 {
					self.playerId = playerId
					self.playerName = name
					fmt.Println("-------CreatePlayer ok:", playerId, name)
					g_httpPlayer = http.NewPlayerRpc(gameAddr, self.accountId)
				} else {
					fmt.Println("-------CreatePlayer:", "创建角色失败")
				}
			})
		}
	})
}

// ------------------------------------------------------------
// -- 方式二：直接登录Tcp游戏服
func (self *loginData) LoginGamesvr_tcp() {
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
	gamesvr.ConnectToSvr(tcp.Addr(self.ip, self.port), netConfig.G_Local_Meta)
}

// ------------------------------------------------------------
// -- 方式三：Gateway网关转接消息
func (self *loginData) LoginGateway() {
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
	gateway.ConnectToSvr(tcp.Addr(self.ip, self.port), netConfig.G_Local_Meta)
}

// ------------------------------------------------------------
func Test_get_gamesvr_list(t *testing.T) {
	loginAddr := netConfig.GetHttpAddr("login", 1)
	http.CallRpc(loginAddr, enum.Rpc_login_get_meta_list, func(buf *common.NetPack) {
		buf.WriteString("") //version
	}, func(backBuf *common.NetPack) {
		cnt := backBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			id := backBuf.ReadInt()
			svrName := backBuf.ReadString()
			outIp := backBuf.ReadString()
			port := backBuf.ReadUInt16()
			playerCnt := backBuf.ReadInt32()
			fmt.Println("-------GamesvrList:", id, svrName, outIp, port, playerCnt)
		}
	})
}
