/***********************************************************************
* @ Gateway
* @ brief
	1、Client先在svr_login完成账户校验，获得accountId、token
	2、再向svr_login要gateway列表，对hash(accountId)决定连哪个gateway

    3、代码生成器可获知，rpc是被哪个模块处理的

	4、RelayPlayerMsg处理的玩家相关rpc（rpc参数是this *TPlayer）
	5、登录之前，游戏服尚无玩家数据，所以“登录、创建”是单独抽离的

* @ optimize
    1、公用的tcp_rpc是单线程的，适合业务逻辑；gateway可改用多线程版，提高转发性能

	轮询的方式也挺不错，gateway收到消息，依次问询后面节点可否处理，并缓存结果信息
	仅玩家上来时会轮询次，后面就固定路由了

* @ author zhoumf
* @ date 2018-3-13
***********************************************************************/
package logic

import (
	"common"
	"generate_out/err"
	"generate_out/rpc/enum"
	"nets/tcp"
	"sync"
)

func Rpc_gateway_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()

	if CheckLoginToken(accountId, token) {
		client.UserPtr = accountId
		AddClientConn(accountId, client)
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Token_verify_err)
	}
}

// RelayPlayerMsg处理的玩家相关rpc（rpc参数是this *TPlayer）
// 登录之前，游戏服尚无玩家数据，所以登录、创建是单独抽离的
func Rpc_gateway_relay_game_login(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if p, ok := GetGameRpc(accountId); ok {
			oldReqKey := req.GetReqKey()
			p.CallRpc(enum.Rpc_game_login, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}
func Rpc_gateway_relay_game_create_player(req, ack *common.NetPack, client *tcp.TCPConn) {
	if accountId, ok := client.UserPtr.(uint32); ok {
		if p, ok := GetGameRpc(accountId); ok {
			oldReqKey := req.GetReqKey()
			p.CallRpc(enum.Rpc_game_create_player, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteBuf(req.LeftBuf())
			}, func(backBuf *common.NetPack) {
				//异步回调，不能直接用ack
				backBuf.SetReqKey(oldReqKey)
				client.WriteMsg(backBuf)
			})
		}
	}
}

// ------------------------------------------------------------
// -- 后台账号验证
var g_login_token sync.Map //<accountId, token>

func Rpc_gateway_login_token(req, ack *common.NetPack, conn *tcp.TCPConn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	gameSvrId := req.ReadInt()

	g_login_token.Store(accountId, token)
	AddPlayer(accountId, gameSvrId) //设置此玩家的game路由
}
func CheckLoginToken(accountId, token uint32) bool {
	if value, ok := g_login_token.Load(accountId); ok {
		return token == value
	}
	return false
}
