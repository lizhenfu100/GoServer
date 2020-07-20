/***********************************************************************
* @ Gateway
* @ brief
	1、Client先在svr_login完成账户校验，获得accountId、token
	2、再向svr_login要gateway列表，对hash(accountId)决定连哪个gateway

    3、代码生成器可获知，rpc是被哪个模块处理的

* @ 跨大区
    、大区配个专门的svr_proxy，统一负责大区间的路由

* @ author zhoumf
* @ date 2018-3-13
***********************************************************************/
package logic

import (
	"common"
	"generate_out/err"
	"github.com/go-redis/redis"
	"netConfig/meta"
	"strconv"
	"sync"
	"time"
)

var g_token sync.Map //<accountId, token>

func Rpc_set_identity(req, ack *common.NetPack, _ common.Conn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	gameSvrId := req.ReadInt()
	g_token.Store(accountId, token)
	AddRouteGame(accountId, gameSvrId) //设置此玩家的game路由
}
func Rpc_check_identity(req, ack *common.NetPack, client common.Conn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()
	ret := err.Token_verify_err
	if v, ok := g_token.Load(accountId); ok && token == v {
		if ret = err.Success; client != nil { //tcp网关
			if p := GetClientConn(accountId); p != nil && p != client {
				p.Close() //防串号
			}
			client.SetUser(accountId)
			AddClientConn(accountId, client)
		}
	}
	ack.WriteUInt16(ret)
}
func Rpc_net_error(req, ack *common.NetPack, conn common.Conn) {
	if accountId, ok := conn.GetUser().(uint32); ok { //玩家断线，且没重连
		if TryDelClientConn(accountId) {
			DelRouteGame(accountId)
		}
	} else if ptr, ok := conn.GetUser().(*meta.Meta); ok {
		if ptr.Module == "game" { //游戏服断开

		}
	}
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

func AddClientConn(aid uint32, conn common.Conn) { g_clients.Store(aid, conn) }
func GetClientConn(aid uint32) common.Conn {
	if v, ok := g_clients.Load(aid); ok {
		return v.(common.Conn)
	}
	return nil
}
func TryDelClientConn(aid uint32) bool {
	if v, ok := g_clients.Load(aid); ok && v.(common.Conn).IsClose() {
		g_clients.Delete(aid)
		return true
	}
	return false
}
