package main

import (
	"common"
	// "conf"
	// "dbmgo"
	"fmt"
	"gamelog"
	"http"
	"netConfig"
	"time"
	//"msg/sdk_msg"
	//"svr_client/api"
	//"svr_game/logic/player"
	"common/net/meta"
	"generate_out/rpc/enum"
	"io"
	nhttp "net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	var v interface{}
	v = uint32(10)
	t := uint32(10)

	addr := "http://192.168.1.11:7033/"
	idx1 := strings.Index(addr, "//") + 2
	idx2 := strings.LastIndex(addr, ":")
	println(addr[idx1:idx2])
	println(addr[idx2+1 : len(addr)-1])

	println(v == t)
	return

	gamelog.InitLogger("client") //初始化日志系统
	gamelog.SetLevel(gamelog.Lv_Debug)
	InitConf()
	netConfig.CreateNetSvr("client", 0)

	// Download()
	test()
	time.Sleep(100 * time.Second)
}
func InitConf() {
	common.G_Csv_Map = map[string]interface{}{
		"conf_net": &meta.G_SvrNets,
	}
	common.LoadAllCsv()
	// for k, v := range netConfig.G_SvrNetCfg {
	// 	fmt.Println(k, v)
	// }
}

func Download() {
	f, err := os.OpenFile("D:/file.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	stat, err := f.Stat() //获取文件状态
	if err != nil {
		panic(err)
	} //把文件指针指到文件末，当然你说为何不直接用 O_APPEND 模式打开，没错是可以。我这里只是试验。
	url := "http://127.0.0.1:7040/table.csv"

	req, _ := nhttp.NewRequest("GET", url, nil)
	req.Header.Set("Range", "bytes="+strconv.FormatInt(stat.Size(), 10)+"-")
	resp, err := nhttp.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}
	println("written: ", written)
}
func test() {
	gameAddr := netConfig.GetHttpAddr("game", -1)
	loginAddr := netConfig.GetHttpAddr("login", -1)
	fmt.Println("---", gameAddr)
	fmt.Println("---", loginAddr)

	//向游戏服请求充值
	// var msg1 sdk_msg.Msg_create_recharge_order_Req
	// msg1.SessionKey = "233xx"
	// msg1.OrderID = "abcdefg233"
	// msg1.Channel = "360"
	// msg1.PlatformEnum = 0
	// msg1.ChargeCsvID = 2
	// buf1, _ := json.Marshal(&msg1)
	// http.PostReq(gameAddr+"create_recharge_order", buf1)

	// //模拟第三方的充值到账
	// sdkAddr := netConfig.GetHttpAddr("sdk", -1)
	// fmt.Println("---", sdkAddr)
	// var msg2 sdk_msg.SDKMsg_recharge_result
	// msg2.OrderID = "abcdefg233"
	// msg2.ThirdOrderID = "zzzzzzzzz"
	// msg2.Channel = "360"
	// msg2.RMB = 233
	// buf2, _ := json.Marshal(&msg2)
	// http.PostReq(sdkAddr+"sdk_recharge_success", buf2)

	time.Sleep(2 * time.Second)

	//向center取游戏服务器列表
	accountName := "zhoumf"
	password := "123"
	accountId := uint32(0)
	gamesvrIp := ""
	gamesvrPort := uint16(0)
	logintoken := uint32(0)

	http.CallRpc(loginAddr, enum.Rpc_login_reg_account, func(buf *common.NetPack) {
		buf.WriteString(accountName)
		buf.WriteString(password)
	}, func(backBuf *common.NetPack) {
		errCode1 := backBuf.ReadInt8()
		fmt.Println("errCode1:", errCode1)
	})

	http.CallRpc(loginAddr, enum.Rpc_login_account_login, func(buf *common.NetPack) {
		buf.WriteInt(1) //gamesvrId
		buf.WriteString(accountName)
		buf.WriteString(password)
	}, func(backBuf *common.NetPack) {
		errCode2 := backBuf.ReadInt8()
		if errCode2 > 0 {
			accountId = backBuf.ReadUInt32()
			gamesvrIp = backBuf.ReadString()
			gamesvrPort = backBuf.ReadUInt16()
			logintoken = backBuf.ReadUInt32()
		} else {
			fmt.Println("errCode2:", errCode2)
		}
	})

	//登录
	playerRpc := http.NewPlayerRpc(gameAddr, 0)
	http.CallRpc(gameAddr, enum.Rpc_game_login, func(buf *common.NetPack) {
		buf.WriteUInt32(accountId)
		buf.WriteUInt32(logintoken)
	}, func(backBuf *common.NetPack) {
		errCode3 := backBuf.ReadInt8()
		if errCode3 > 0 {
			playerRpc.PlayerId = backBuf.ReadUInt32()
		} else if errCode3 == -2 {
			//创建新角色
			http.CallRpc(gameAddr, enum.Rpc_game_player_create, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteString("zhoumf")
			}, func(backBuf *common.NetPack) {
				playerRpc.PlayerId = backBuf.ReadUInt32()
			})
		} else {
			fmt.Println("errCode3:", errCode3)
		}
	})

	//登出
	// playerRpc.CallRpc(enum.Rpc_game_logout, func(buf *common.NetPack) {
	// }, func(backBuf *common.NetPack) {
	// })

	// time.Sleep(2 * time.Second)
	// //直接发给战斗服
	// msg := common.NewNetPackCap(32)
	// msg.SetRpc("rpc_echo")
	// msg.WriteString("--- zhoumf 233 --- ")
	// api.SendToBattle(1, msg)
	// api.SendToBattle(2, msg)

	time.Sleep(2 * time.Hour)
}
