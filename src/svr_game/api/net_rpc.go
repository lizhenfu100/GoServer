package api

import (
	"common"
	"common/net/meta"
	"encoding/json"
	"http"
	"math/rand"
	"netConfig"
	"tcp"
)

//Notice：TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
var (
	g_cache_cross_conn = make(map[int]*tcp.TCPConn)
	g_cache_login_addr string
	g_cache_sdk_addr   string
)

// ------------------------------------------------------------
//! cross
func CallRpcCross(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	ids := meta.GetModuleIDs("cross")
	id := ids[rand.Intn(len(ids))] //随机一个节点

	conn, _ := g_cache_cross_conn[id]
	if conn == nil || conn.IsClose() {
		conn = netConfig.GetTcpConn("cross", id)
		g_cache_cross_conn[id] = conn
	}
	conn.CallRpc(rid, sendFun, recvFun)
}

// ------------------------------------------------------------
//! login
func CallRpcLogin(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if g_cache_login_addr == "" {
		g_cache_login_addr = netConfig.GetHttpAddr("login", -1)
	}
	http.CallRpc(g_cache_login_addr, rid, sendFun, recvFun)
}

// ------------------------------------------------------------
//! sdk
func SendToSdk(strKey string, pMsg interface{}) []byte { // strKey = "create_recharge_order"
	if g_cache_sdk_addr == "" {
		g_cache_sdk_addr = netConfig.GetHttpAddr("sdk", -1)
	}

	buf, _ := json.Marshal(pMsg)
	url := g_cache_sdk_addr + strKey

	return http.PostReq(url, buf)
}
