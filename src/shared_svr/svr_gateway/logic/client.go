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
	"common/tool/wechat"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"netConfig"
	"netConfig/meta"
	"nets/tcp"
	"sync"
	"time"
)

const kDBRoute = "route"

// ------------------------------------------------------------
// -- 后台账号验证
var g_token sync.Map //<accountId, token>

func Rpc_gateway_login_token(req *common.NetPack) {
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
var g_route_game sync.Map //accountId-gameSvrId
type RouteGame struct {
	AccountId uint32 `bson:"_id"`
	SvrId     int
	Time      int64
}

func InitRouteGame() {
	timenow, list := time.Now().Unix(), []RouteGame{}
	if dbmgo.FindAll(kDBRoute, bson.M{"time": bson.M{"$gt": timenow - 2*3600}}, &list) == nil {
		for _, v := range list {
			if netConfig.HashGatewayID(v.AccountId) == meta.G_Local.SvrID { //应连本节点的玩家
				g_route_game.Store(v.AccountId, v.SvrId)
			}
		}
	}
}
func AddRouteGame(aid uint32, svrId int) {
	g_route_game.Store(aid, svrId)
	v := RouteGame{aid, svrId, time.Now().Unix()}
	dbmgo.UpsertId(kDBRoute, aid, &v)
}
func DelRouteGame(aid uint32) { g_route_game.Delete(aid) }

var (
	g_clients = map[uint32]*tcp.TCPConn{}
	g_mutex   sync.RWMutex
)

func TryDelClientConn(aid uint32) bool {
	g_mutex.Lock()
	if v, ok := g_clients[aid]; ok && v.IsClose() {
		delete(g_clients, aid)
		g_mutex.Unlock()
		return true
	}
	g_mutex.Unlock()
	return false
}
func AddClientConn(aid uint32, conn *tcp.TCPConn) {
	g_mutex.Lock()
	g_clients[aid] = conn
	g_mutex.Unlock()
}
func GetClientConn(aid uint32) *tcp.TCPConn {
	g_mutex.RLock()
	ret := g_clients[aid]
	g_mutex.RUnlock()
	return ret
}

// ------------------------------------------------------------
// 辅助函数
func GetModuleRpc(module string, aid uint32) (netConfig.Rpc, bool) {
	if module == "game" {
		return GetGameRpc(aid)
	} else { // 其它节点无状态的，AccountId取模得节点id
		ids := meta.GetModuleIDs(module, meta.G_Local.Version)
		if length := uint32(len(ids)); length > 0 {
			id := ids[aid%length]
			return netConfig.GetRpc(module, id)
		}
	}
	return nil, false
}
func GetGameRpc(aid uint32) (netConfig.Rpc, bool) {
	if v, ok := g_route_game.Load(aid); ok {
		return netConfig.GetGameRpc(v.(int))
	} else {
		v := RouteGame{}
		if ok, _ := dbmgo.Find(kDBRoute, "_id", aid, &v); ok {
			g_route_game.Store(aid, v.SvrId)
			return netConfig.GetGameRpc(v.SvrId)
		}
	}
	wechat.SendMsg(fmt.Sprintf("lose game route: %d", aid))
	return nil, false
}
