package console

import (
	"common"
	"common/assert"
	"encoding/json"
	"fmt"
	"gamelog"
	"netConfig/meta"
	zk "shared_svr/zookeeper/logic"
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
func _Rpc_gm_cmd(req, ack *common.NetPack, _ common.Conn) {
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
	"setmeta": cmd_setmeta, //不含空格{"Module":"game","SvrID":1,"Closed":true}
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
	jsonData := common.S2B(args[0])
	pNew := &meta.Meta{}
	json.Unmarshal(jsonData, pNew)
	if p := meta.GetMeta(pNew.Module, pNew.SvrID); p != nil {
		json.Unmarshal(jsonData, p)
		pNew = p
	} else {
		meta.AddMeta(pNew)
	}
	_on_edit_zk(pNew) //GM编辑zookeeper，控制集群路由关系

	b, _ := json.MarshalIndent(pNew, "", "     ")
	return common.B2S(b)
}
func _on_edit_zk(p *meta.Meta) {
	if meta.G_Local.Module == meta.Zookeeper {
		if p.Closed {
			zk.OnMetaDel(p)
		} else {
			zk.OnMetaAdd(p)
		}
	}
}
