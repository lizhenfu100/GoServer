/***********************************************************************
* @ 联机玩法
* @ brief
    · 局域网发现
        · 优化后的版本尚需外网验证
    · Mirror广域网改造
        · 还有优化空间（接管Mirror通信模块Transport）
    · 广域网组队，适配各类组队方式
        · 邀请码（开发中）
        · 后台匹配
    · 外网通信优化
        · 可靠udp，并包优化
        · 防抖动
		· 丢包方案（比较专业，需持续投入）
    · 项目源码的联机改造
        · Network框架的注意事项，大体使用思路（粗略即可）
        · bug示例+分析，附上源码

* @ NAT穿透
    两个客户端分别用udp发送自己的内网ip+端口号到外网的服务器上，
    这样外网的服务器就可以得到两个客户端的内部ip+端口号和外部ip+端口号
    然后外网服务器吧客户端1的发给2，2的发给1，
    两边得到的是对面的内&外ip+端口号，两边同时用udp去连(且发个包)，有一边连得上就打得通
    连不上那说明有一边是对称型，需要中继
		、完全圆锥形
			· 一个内部对一个外部
			· 任意主机都能发包至内网
		、受限圆锥形
			· 内网须先发包给对方
		、端口受限圆锥形
			· 对方的端口也须固定
		、对称
			· 内网须先发包给对方
			· 一个内部对应多个外部

* @ author zhoumf
* @ date 2020-5-18
************************************************************************/
package player

import (
	"common"
	"gamelog"
)

func Rpc_relay_to_others(req, ack *common.NetPack, this *Player) {
	for _, v := range this.room.list {
		if v != this {
			v.Conn.WriteMsg(req)
		}
	}
}
func Rpc_relay_to_other(req, ack *common.NetPack, _ common.Conn) {
	pid := req.ReadUInt32()
	rid := req.ReadUInt16()
	args := req.LeftBuf()
	if v := FindPlayer(pid); v != nil {
		gamelog.Debug("relay msg: %d", rid)
		v.Conn.CallRpc(rid, func(buf *common.NetPack) {
			buf.WriteBuf(args)
		}, nil)
	} else {
		gamelog.Error("relay_to_other: rid(%d) pid(%d)", rid, pid)
	}
}
