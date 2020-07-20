/***********************************************************************
* @ 发邀请码建房
* @ brief
    · 房主建房后，回复邀请码
	· 微信/QQ通知好友，对方输入“邀请码”加入对应房间

* @ author zhoumf
* @ date 2020-5-15
***********************************************************************/
package invite_room

import (
	"common"
	"common/std/random"
	"gamelog"
	"generate_out/rpc/enum"
	"shared_svr/svr_relay/player"
	"sync"
)

var g_c2p sync.Map //<string, *Player>
var g_p2c sync.Map //<*Player, string>

func Rpc_relay_c2p(req, ack *common.NetPack, this *player.Player) {
	code := req.ReadString()
	if v, ok := g_c2p.Load(code); ok {
		if pHost := v.(*player.Player); !pHost.Conn.IsClose() {
			ip, port := pHost.ClientAddr()
			ack.WriteUInt32(pHost.Pid)
			ack.WriteString(ip)
			ack.WriteUInt16(port)
			// 同时通知主机、客机，都向对方发包（受限圆锥形NAT）
			pHost.Conn.CallRpc(enum.Rpc_client_notify_addr, func(buf *common.NetPack) {
				ip, port := this.ClientAddr()
				buf.WriteUInt32(this.Pid)
				buf.WriteString(ip)
				buf.WriteUInt16(port)
			}, nil)
		} else {
			gamelog.Error("Offline: %v", pHost.Pid)
			g_c2p.Delete(code)
			g_p2c.Delete(pHost)
		}
	}
}
func Rpc_relay_p2c(req, ack *common.NetPack, this *player.Player) {
	var code string
	if v, ok := g_p2c.Load(this); ok {
		code = v.(string)
	} else {
		code = "zzz233"
		g_c2p.Store(code, this)
		g_p2c.Store(this, code)
	}
	ack.WriteString(code)
	gamelog.Track("pid(%d) code(%s)", this.Pid, code)
}

func newCode(ptr *player.Player) string {
	code := random.String(5)
	if _, ok := g_c2p.Load(code); ok {
		return newCode(ptr)
	} else {
		g_c2p.Store(code, ptr)
		g_p2c.Store(ptr, code)
	}
	return code
}
