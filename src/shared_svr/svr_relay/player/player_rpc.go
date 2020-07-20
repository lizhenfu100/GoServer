package player

import (
	"common"
	"common/service"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net"
	"netConfig"
	"nets/rpc"
	"sync"
	"sync/atomic"
)

type PlayerRpc func(r, w *common.NetPack, this *Player)

var G_PlayerHandleFunc [enum.RpcEnumCnt]PlayerRpc

func RegPlayerRpc(list map[uint16]PlayerRpc) {
	for k, v := range list {
		G_PlayerHandleFunc[k] = v
	}
	rpc.RegHandlePlayerRpc(_PlayerRpcTcp) //tcp 直连
}
func _PlayerRpcTcp(req, ack *common.NetPack, conn common.Conn) bool {
	rpcId := req.GetMsgId()
	if msgFunc := G_PlayerHandleFunc[rpcId]; msgFunc != nil {
		if p, ok := conn.GetUser().(*Player); ok {
			atomic.StoreUint32(&p._idleMin, 0)
			msgFunc(req, ack, p)
		} else {
			gamelog.Error("offline")
		}
		return true
	}
	return false
}

// ------------------------------------------------------------
var (
	g_cache   sync.Map //<pid, *Player>
	G_Service *service.ServiceVec
)

type Player struct {
	Pid      uint32
	_idleMin uint32 //每次收到消息时归零
	Conn     common.Conn
	room     *Room
}

func init() {
	G_Service = service.NewServiceVec(_Service_Check_AFK, 60*1000)
}
func _Service_Check_AFK(ptr interface{}) {
	if p, ok := ptr.(*Player); ok {
		if atomic.AddUint32(&p._idleMin, 1) > 5 {
			G_Service.UnRegister(p)
			p.Logout()
			g_cache.Delete(p)
		}
	}
}
func FindPlayer(pid uint32) *Player {
	if v, ok := g_cache.Load(pid); ok {
		return v.(*Player)
	}
	return nil
}
func Rpc_check_identity(req, ack *common.NetPack, conn common.Conn) {
	pid := req.ReadUInt32()
	code := req.ReadString()
	gamelog.Track("Login: %d", pid)
	p := FindPlayer(pid)
	if p == nil {
		p = &Player{Pid: pid}
		g_cache.Store(pid, p)
		G_Service.Register(p)
	}
	if p.Conn != nil && p.Conn != conn {
		p.Conn.Close() //防串号
	}
	p.Conn = conn
	p.room = nil //测试代码
	conn.SetUser(p)
	atomic.StoreUint32(&p._idleMin, 0)

	if ClientHash(code) {
		ip, port := p.ClientAddr()
		ack.WriteUInt16(err.Success)
		ack.WriteString(ip)
		ack.WriteUInt16(port)
	} else {
		ack.WriteUInt16(err.Is_forbidden)
	}
}
func (p *Player) ClientAddr() (string, uint16) {
	if p.Conn != nil {
		ip, port := common.ParseAddr(p.Conn.(net.Conn).RemoteAddr().String())
		gamelog.Track("%s:%d %d", ip, port, p.Pid)
		return ip, port
	}
	return "", 0
}
func (p *Player) Logout() {
}

// ------------------------------------------------------------
var _clientHash sync.Map

func ClientHash(v string) (ret bool) {
	ret = true
	if _, ok := _clientHash.Load(v); !ok {
		if p, ok := netConfig.GetRpcRand("gm"); ok { //【只国内转发服配gm】
			p.CallRpc(enum.Rpc_gm_client_hash, func(buf *common.NetPack) {
				buf.WriteString(v)
			}, func(recvbuf *common.NetPack) {
				if ret = recvbuf.ReadBool(); ret {
					_clientHash.Store(v, true)
				}
			})
		}
	}
	return
}
