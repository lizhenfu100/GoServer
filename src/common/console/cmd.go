package console

import (
	"common"
	"common/assert"
	"encoding/json"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"nets/tcp"
	"strconv"
	"strings"
)

type cmdFunc func(args []string) string

func RegCmd(key string, f cmdFunc) {
	if _, ok := g_cmds[key]; ok {
		assert.True(false)
		gamelog.Error("RegCmd repeat: " + key)
		return
	}
	g_cmds[key] = f
}
func _Rpc_gm_cmd1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_gm_cmd2(req, ack) }
func _Rpc_gm_cmd2(req, ack *common.NetPack) {
	cmd := req.ReadString()
	args_ := req.ReadString()
	args := strings.Split(args_, " ")
	defer func() {
		if r := recover(); r != nil {
			ack.WriteString(fmt.Sprintf("%v", r))
		}
	}()
	if cmd, ok := g_cmds[cmd]; ok {
		ack.WriteString(cmd(args))
	} else {
		ack.WriteString("none cmd")
	}
}

var g_cmds = map[string]cmdFunc{ //Notice：注意下列函数的线程安全性
	"loglv":   cmd_logLv,
	"metas":   cmd_metas,
	"setmeta": cmd_setmeta,
}

// ------------------------------------------------------------
//! 命令行函数
func cmd_logLv(args []string) string {
	lv, _ := strconv.Atoi(args[0])
	gamelog.SetLevel(lv)
	return fmt.Sprintf("SetLogLv: %d", lv)
}
func cmd_metas(args []string) string {
	vs := meta.GetMetas(args[0], "")
	b, _ := json.MarshalIndent(vs, "", "     ")
	return common.B2S(b)
}
func cmd_setmeta(args []string) string {
	jsonData := common.S2B(args[0]) //不含空格{"Module":"game","SvrID":1,"IP":"192.168.1.111"}
	pNew := &meta.Meta{}
	json.Unmarshal(jsonData, pNew)
	if p := meta.GetMeta(pNew.Module, pNew.SvrID); p != nil {
		json.Unmarshal(jsonData, p)
		pNew = p
	} else {
		meta.AddMeta(pNew)
	}
	b, _ := json.MarshalIndent(pNew, "", "     ")
	return common.B2S(b)
}
