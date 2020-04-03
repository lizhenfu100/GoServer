/***********************************************************************
* @ 选取节点id
* @ brief
	1、取模方式的模块，不支持动态增删
		、gateway、friend、分流game
		、center比较特殊，它本身是无状态的
			玩家分到错误节点也能重新从数据库读到自己数据

* @ FIXME：一些hash取模定位的节点，依赖了节点总数；节点陆续连接，中途玩家就上来通信，会分配至错误节点
	gateway，带状态的，一旦分配错误，影响很大
	friend，若联了不同的db_friend，会找不到数据
	center，联的同个db，影响不大，只是缓存不友好

* @ author zhoumf
* @ date 2019-1-18
***********************************************************************/
package netConfig

import (
	"common"
	"common/std/hash"
	"gamelog"
	"generate_out/rpc/enum"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
	"sync"
)

// ------------------------------------------------------------
// 统一接口
type Rpc interface {
	IsClose() bool
	CallRpc(rid uint16, sendFun, recvFun func(*common.NetPack))
	CallRpcSafe(rid uint16, sendFun, recvFun func(*common.NetPack))
}
type rpcHttp struct{ *meta.Meta }

func (p *rpcHttp) IsClose() bool { return p.IsClosed }
func (p *rpcHttp) CallRpc(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	addr := http.Addr(p.IP, p.HttpPort)
	http.CallRpc(addr, rid, sendFun, recvFun)
}
func (p *rpcHttp) CallRpcSafe(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	p.CallRpc(rid, sendFun, recvFun)
}

//interface无法"!= nil"判别有效
//var p netConfig.Rpc
//p = netConfig.GetTcpConn("game", 2)
//fmt.Println(p == nil, p) //false nil
func GetRpc1(p *meta.Meta) (Rpc, bool) {
	if p.HttpPort > 0 {
		return &rpcHttp{p}, true
	} else {
		ret := GetTcpConn(p.Module, p.SvrID)
		return ret, ret != nil
	}
}
func GetRpc2(module string, svrId int) (Rpc, bool) {
	if p := meta.GetMeta(module, svrId); p == nil {
		return nil, false
	} else {
		return GetRpc1(p)
	}
}

// ------------------------------------------------------------
//! 无状态节点，随机取 -- login friend sdk save gm
func GetRpcRand(module string) (Rpc, bool) {
	if p := meta.GetByRand(module); p != nil {
		return GetRpc1(p)
	}
	return nil, false
}

// ------------------------------------------------------------
//! center http -- 账号名hash取模
func HashCenterID(key string) int {
	list := meta.GetMetas("center", meta.G_Local.Version)
	if n := uint32(len(list)); n > 0 {
		return list[hash.StrHash(key)%n].SvrID
	}
	return -1
}
func CallRpcCenter(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if addr := GetHttpAddr("center", svrId); addr != "" {
		http.CallRpc(addr, rid, sendFun, recvFun)
	} else {
		gamelog.Error("center nil: svrId(%d) rpcId(%d)", svrId, rid)
	}
}

// ------------------------------------------------------------
//! gateway tcp|http -- 账号hash取模
var g_cache_gate sync.Map //<int, Rpc>

func HashGatewayID(accountId uint32) int { //TODO：用一致性hash，取模方式gateway无法动态扩展
	list := meta.GetMetas("gateway", meta.G_Local.Version)
	if n := uint32(len(list)); n > 0 {
		return list[accountId%n].SvrID
	}
	return -1
}
func CallRpcGateway(accountId uint32, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	svrId := HashGatewayID(accountId)
	if p, ok := GetGatewayRpc(svrId); ok {
		p.CallRpcSafe(enum.Rpc_gateway_relay, func(buf *common.NetPack) {
			buf.WriteUInt16(rid)
			buf.WriteUInt32(accountId)
			sendFun(buf)
		}, recvFun)
	} else {
		gamelog.Error("gateway nil: svrId(%d) rpcId(%d)", svrId, rid)
	}
}
func GetGatewayRpc(svrId int) (ret Rpc, ok bool) {
	var v interface{}
	if v, ok = g_cache_gate.Load(svrId); !ok {
		if ret, ok = GetRpc2("gateway", svrId); ok {
			g_cache_gate.Store(svrId, ret)
		}
	} else if ret, ok = v.(Rpc); !ok || ret.IsClose() {
		if ret, ok = GetRpc2("gateway", svrId); ok {
			g_cache_gate.Store(svrId, ret)
		}
	}
	return
}

// ------------------------------------------------------------
//! game tcp|http
var g_cache_game sync.Map //<int, Rpc>

func GetGameRpc(svrId int) (ret Rpc, ok bool) {
	var v interface{}
	if v, ok = g_cache_game.Load(svrId); !ok {
		if ret, ok = GetRpc2("game", svrId); ok {
			g_cache_game.Store(svrId, ret)
		}
	} else if ret, ok = v.(Rpc); !ok || ret.IsClose() {
		if ret, ok = GetRpc2("game", svrId); ok {
			g_cache_game.Store(svrId, ret)
		}
	}
	return
}

// ------------------------------------------------------------
//! cross & battle -- 非线程安全
var g_cache_cross = make(map[int]*tcp.TCPConn)
var g_cache_battle = make(map[int]*tcp.TCPConn)

func CallRpcCross(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if p := meta.GetByRand("cross"); p != nil {
		if conn := GetCrossConn(p.SvrID); conn != nil {
			conn.CallRpc(rid, sendFun, recvFun)
			return
		}
	}
	gamelog.Error("cross nil: rpcId(%d)", rid)
}
func CallRpcBattle(svrId int, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	if conn := GetBattleConn(svrId); conn != nil {
		conn.CallRpc(rid, sendFun, recvFun)
	} else {
		gamelog.Error("battle nil: svrId(%d) rpcId(%d)", svrId, rid)
	}
}
func GetCrossConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_cross[svrId]
	if conn == nil || conn.IsClose() {
		conn = GetTcpConn("cross", svrId)
		g_cache_cross[svrId] = conn
	}
	return conn
}
func GetBattleConn(svrId int) *tcp.TCPConn {
	conn, _ := g_cache_battle[svrId]
	if conn == nil || conn.IsClose() {
		conn = GetTcpConn("battle", svrId)
		g_cache_battle[svrId] = conn
	}
	return conn
}
