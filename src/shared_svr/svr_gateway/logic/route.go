/***********************************************************************
* @ Gateway
* @ brief
	1、Client先在svr_login完成账户校验，获得accountId、token
	2、再向svr_login要gateway列表，对hash(accountId)决定连哪个gateway

    3、代码生成器可获知，rpc是被哪个模块处理的

* @ optimize
    1、公用的tcp_rpc是单线程的，适合业务逻辑；gateway可改用多线程版，提高转发性能

* @ author zhoumf
* @ date 2018-3-13
***********************************************************************/
package logic

import (
	"common"
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"net/http/httputil"
	"netConfig/meta"
	"nets/tcp"
	"strconv"
	"sync"
	"time"
)

var g_token sync.Map //<accountId, token>

func Rpc_set_identity(req *common.NetPack) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	gameSvrId := req.ReadInt()
	g_token.Store(accountId, token)
	AddRouteGame(accountId, gameSvrId) //设置此玩家的game路由
}
func CheckToken(accountId, token uint32) bool {
	if value, ok := g_token.Load(accountId); ok {
		return token == value
	}
	return false
}

// ------------------------------------------------------------
// 与玩家相关的网络连接
var (
	g_clients    sync.Map //<aid, *TCPConn>
	g_route_game sync.Map //<aid, gameId>
	g_redis      = redis.NewClient(&redis.Options{Addr: ":6379"})
)

func init() {
	if _, e := g_redis.Ping().Result(); e != nil {
		panic(e)
	}
}
func AddRouteGame(aid uint32, svrId int) {
	g_route_game.Store(aid, svrId)
	g_redis.Set(strconv.FormatInt(int64(aid), 10), svrId, 24*time.Hour)
}
func GetGameId(aid uint32) int {
	if v, ok := g_route_game.Load(aid); ok {
		return v.(int)
	} else if v, e := g_redis.Get(strconv.FormatInt(int64(aid), 10)).Int(); e == nil {
		g_route_game.Store(aid, v)
		return v
	}
	return -1
}
func DelRouteGame(aid uint32) { g_route_game.Delete(aid) }

func AddClientConn(aid uint32, conn *tcp.TCPConn) { g_clients.Store(aid, conn) }
func GetClientConn(aid uint32) *tcp.TCPConn {
	if v, ok := g_clients.Load(aid); ok {
		return v.(*tcp.TCPConn)
	}
	return nil
}
func TryDelClientConn(aid uint32) bool {
	if v, ok := g_clients.Load(aid); ok && v.(*tcp.TCPConn).IsClose() {
		g_clients.Delete(aid)
		return true
	}
	return false
}

// ------------------------------------------------------------
// 转发http
var _sdk http.Handler

func relaySdk(w http.ResponseWriter, r *http.Request) { _sdk.ServeHTTP(w, r) }
func init() {
	_sdk = &httputil.ReverseProxy{Director: func(r *http.Request) {
		if p := meta.GetByRand("sdk"); p != nil {
			r.URL.Scheme = "http"
			r.URL.Host = fmt.Sprintf("%s:%d", p.IP, p.HttpPort)
		}
	}}
	http.HandleFunc("/pre_buy_request", relaySdk)
	http.HandleFunc("/query_order", relaySdk)
	http.HandleFunc("/confirm_order", relaySdk)
	http.HandleFunc("/query_order_unfinished", relaySdk)
}
