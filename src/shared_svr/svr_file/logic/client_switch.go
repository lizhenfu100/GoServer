package logic

import (
	"net/http"
	"strconv"
	"sync"
)

// ------------------------------------------------------------
// -- 动态开关客户端debug日志
var g_macs sync.Map

func Client_debug_log(args []string) string { //ClientLog ed6a844be9d7a607 1
	mac := args[0]
	open, _ := strconv.Atoi(args[1])
	if open > 0 {
		g_macs.Store(mac, nil)
		return "debug open"
	} else {
		g_macs.Delete(mac)
		return "debug close"
	}
}
func Http_client_debug_log(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if _, ok := g_macs.Load(q.Get("mac")); ok {
		w.Write([]byte{1})
	} else {
		w.Write([]byte{0})
	}
}
