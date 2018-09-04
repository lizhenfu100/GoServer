package main

import (
	"bytes"
	"common"
	"common/file"
	"common/timer"
	"flag"
	"fmt"
	"gamelog"
	"generate_out/rpc/enum"
	"http"
	"os"
	"strings"
	"svr_sdk/msg"
	"text/template"
	"time"
)

var (
	//命令行标志，须出现于参数之前（否则，该标志会被解析为位置参数）
	_g, _s, _ip string
	_port       int
	g_addr      string
)

func init() {
	flag.StringVar(&_g, "g", "", "查询订单信息")
	flag.StringVar(&_s, "s", "", "修改失败订单")
	flag.StringVar(&_ip, "ip", "120.78.152.152", "ip")
	flag.IntVar(&_port, "port", 7002, "port")
}

func main() {
	gamelog.InitLogger("Order")
	flag.Parse() //内部获取了所有参数：os.Args[1:]

	g_addr = http.Addr(_ip, uint16(_port))

	//fmt.Println(g_addr, flag.Args())
	//_g = "0091805016400525 0091805016400382 0091805016395545"

	GetOrderInfo(strings.Split(_g, " "))
	OrderSuccess(strings.Split(_s, " "))

	time.Sleep(time.Hour)
}

// --------------------------------------------------------------------------
//
func OrderSuccess(orderIds []string) {
	if len(orderIds) > 0 && orderIds[0] != "" {
		gamelog.Debug("set: %d %v", len(orderIds), orderIds)

		http.CallRpc(g_addr, enum.Rpc_order_success, func(buf *common.NetPack) {
			buf.WriteUInt16(uint16(len(orderIds)))
			for _, v := range orderIds {
				buf.WriteString(v)
			}
		}, func(recvBuf *common.NetPack) {
			if err := recvBuf.ReadString(); err != "" {
				gamelog.Error(err)
				fmt.Println(err)
			}
		})
	}
}
func GetOrderInfo(orderIds []string) {
	if len(orderIds) > 0 && orderIds[0] != "" {
		gamelog.Debug("get: %d %v", len(orderIds), orderIds)

		http.CallRpc(g_addr, enum.Rpc_order_info, func(buf *common.NetPack) {
			buf.WriteUInt16(uint16(len(orderIds)))
			for _, v := range orderIds {
				buf.WriteString(v)
			}
		}, func(recvBuf *common.NetPack) {
			cnt := recvBuf.ReadUInt16()
			vec := make([]msg.TOrderInfo, cnt)
			for i := uint16(0); i < cnt; i++ {
				vec[i].Order_id = orderIds[i]
				if ok := recvBuf.ReadInt8(); ok > 0 {
					vec[i].Third_order_id = recvBuf.ReadString()
					vec[i].Third_account = recvBuf.ReadString()
					vec[i].Item_name = recvBuf.ReadString()
					vec[i].Item_count = recvBuf.ReadInt()
					vec[i].Total_price = recvBuf.ReadInt()
					vec[i].Extra = timer.Time2Str(recvBuf.ReadInt64()) //临时用于时间戳转日期
					vec[i].Status = recvBuf.ReadInt()
					vec[i].Can_send = recvBuf.ReadInt()
				} else {
					vec[i].Order_id += " ==> 查无此号"
					gamelog.Error(vec[i].Order_id)
					fmt.Println(vec[i].Order_id)
				}
			}
			makeFile(vec)
		})
	}
}

// --------------------------------------------------------------------------
// 查到的订单，输出成文本
const K_Out_Template = `{{range .}}
{
	游戏订单号：  {{.Order_id}}
	平台订单号：  {{.Third_order_id}}
	账号ID：      {{.Third_account}}
	商品名称：    {{.Item_name}}
	商品数量：    {{.Item_count}}
	总价：        {{.Total_price}}
	日期：        {{.Extra}}
	订单状态：    {{.Status}}
	能否发货：    {{.Can_send}}
}
{{end}}
`

func makeFile(vec []msg.TOrderInfo) {
	filename := time.Now().Format("20060102_150405") + ".log"
	tpl, err := template.New(filename).Parse(K_Out_Template)
	if err != nil {
		panic(err.Error())
		return
	}
	var bf bytes.Buffer
	if err = tpl.Execute(&bf, vec); err != nil {
		panic(err.Error())
		return
	}
	f, err := file.CreateFile("Order/", filename, os.O_WRONLY|os.O_APPEND)
	if err != nil {
		panic(err.Error())
		return
	}
	defer f.Close()
	f.Write(bf.Bytes())

	if len(vec) < 5 {
		fmt.Println(bf.String())
	}
}
