package test

import (
	"common"
	"conf"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"io/ioutil"
	"netConfig"
	"nets/http"
	"nets/tcp"
	"os"
	"strconv"
	_ "svr_client/test/init"
	"testing"
	"time"
)

type loginData struct {
	account    string
	passwd     string
	typ        string
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
		typ:        "email",
		passwd:     "123123",
		playerName: "test",
	}
	g_centerAddr = netConfig.GetHttpAddr("center", 1)
	g_loginList  = []string{
		"http://192.168.1.111:7030",
		"http://52.14.1.205:7030",    //1 北美
		"http://13.229.215.168:7030", //2 亚洲
		"http://18.185.80.202:7030",  //3 欧洲
		"http://54.94.211.178:7030",  //4 南美
		"http://39.96.196.250:7030",  //5 华北 大家饿
		"http://47.106.35.74:7030",   //6 华南 大家饿
		//"http://39.97.111.110:7030", //5 华北 元气
		//"http://39.108.87.225:7030", //6 华南 元气
	}
	g_loginAddr = g_loginList[6]
	g_version   = ""
)

// go test -v ./src/svr_client/test/login_test.go
func Test_login_init(t *testing.T) {
	var saveBuf []byte
	if f, e := os.Open("D:/diner_svr/" + "5307437.save"); e == nil {
		saveBuf, _ = ioutil.ReadAll(f)
		f.Close()
	} else {
		panic("open file err")
	}

	for i := 1; i <= 100; i++ {
		g_loginData.account = fmt.Sprintf("chillyroomtest%d@test.com", i)
		g_loginData.passwd = "123456"
		tTest_account_reg(t)
		tTest_account_login(t)

		url := fmt.Sprintf("http://47.106.35.74:7040/whitelist_add/?passwd=%s&val=%d", conf.GM_Passwd, g_loginData.accountId)
		http.Client.Get(url)

		time.Sleep(time.Second * 3)

		http.CallRpc("http://47.106.35.74:7090", enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
			buf.WriteString(strconv.Itoa(int(g_loginData.accountId)))
			buf.WriteString("Android")
			buf.WriteString(g_loginData.account)
			buf.WriteString("") //sign
			buf.WriteString("") //extra
			buf.WriteLenBuf(saveBuf)
			buf.WriteString("")
		}, func(backBuf *common.NetPack) {
			errCode := backBuf.ReadUInt16()
			fmt.Println("------------: ", g_loginData.accountId, errCode, len(saveBuf))
		})
	}
}

// ------------------------------------------------------------
func tTest_get_gamesvr_list(t *testing.T) {
	http.CallRpc(g_loginAddr, enum.Rpc_login_get_game_list, func(buf *common.NetPack) {
		buf.WriteString(g_version)
	}, func(backBuf *common.NetPack) {
		for cnt, i := backBuf.ReadByte(), byte(0); i < cnt; i++ {
			id := backBuf.ReadInt()
			outIp := backBuf.ReadString()
			port := backBuf.ReadUInt16()
			svrName := backBuf.ReadString()
			onlineCnt := backBuf.ReadInt32()
			fmt.Println("-------GamesvrList:", id, svrName, outIp, port, onlineCnt)
		}
	})
}
func tTest_player_gameinfo_addr(t *testing.T) {
	http.CallRpc(g_centerAddr, enum.Rpc_center_player_addr2, func(buf *common.NetPack) {
		buf.WriteString(g_loginData.account)
		buf.WriteString("email")
		buf.WriteString(g_gameName)
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
	http.CallRpc(g_loginAddr, enum.Rpc_login_to_center_by_str, func(buf *common.NetPack) {
		buf.WriteUInt16(enum.Rpc_center_account_reg2)
		buf.WriteString(g_loginData.account)
		buf.WriteString(g_loginData.passwd)
		buf.WriteString(g_loginData.typ)
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
	gameAddr := http.Addr(self.ip, self.port)
	http.CallRpc(gameAddr, enum.Rpc_check_identity, func(buf *common.NetPack) {
		buf.WriteUInt32(self.accountId)
		buf.WriteUInt32(self.token)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		fmt.Println("-------GameLogin errCode:", errCode)
		if errCode == err.Success {
			self.playerId = backBuf.ReadUInt32()
			self.playerName = backBuf.ReadString()
			fmt.Println("-------GameLogin ok:", self.accountId, self.playerId, self.playerName)
		}
	})
}
func (self *loginData) LoginGamesvr_tcp() { //方式二：直接登录Tcp游戏服
	gamesvr := new(tcp.TCPClient)
	gamesvr.Connect(tcp.Addr(self.ip, self.port), func(conn *tcp.TCPConn) {
		conn.CallEx(enum.Rpc_check_identity, func(buf *common.NetPack) {
			buf.WriteUInt32(self.accountId)
			buf.WriteUInt32(self.token)
		}, func(backBuf *common.NetPack) {
			errCode := backBuf.ReadUInt16()
			fmt.Println("-------GameLogin errCode:", errCode)
			if errCode == err.Success {
				self.playerId = backBuf.ReadUInt32()
				self.playerName = backBuf.ReadString()
				fmt.Println("-------GameLogin ok:", self.playerId, self.playerName)
			}
		})
	})
}
func (self *loginData) LoginGateway() { //方式三：Gateway网关转接消息
	gateway := new(tcp.TCPClient)
	gateway.Connect(tcp.Addr(self.ip, self.port), func(conn *tcp.TCPConn) {
		conn.CallEx(enum.Rpc_check_identity, func(buf *common.NetPack) {
			buf.WriteUInt32(self.accountId)
			buf.WriteUInt32(self.token)
		}, func(backBuf *common.NetPack) {
			errCode := backBuf.ReadUInt16()
			fmt.Println("-------GatewayLogin errCode:", errCode)
		})
	})
}

// ------------------------------------------------------------
// --
