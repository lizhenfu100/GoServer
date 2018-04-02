package logic

import (
	"common"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"sync"
	"sync/atomic"
)

var g_login_token uint32

func Rpc_login_account_login(req, ack *common.NetPack) { //req: 游戏区服id，游戏名，账号，密码
	gameSvrId := req.ReadInt() //玩家手选的区服

	//临时读取buffer数据
	oldPos := req.ReadPos
	/*gameName :=*/ req.ReadString()
	account := req.ReadString()
	req.ReadPos = oldPos

	pMeta := meta.GetMeta("game", gameSvrId)
	if pMeta.Port() <= 0 {
		ack.WriteInt8(-100) //invalid_svrid
		return
	}
	//Notice: 这里必须是同步调用，CallRpc的回调里才能有效使用ack
	//由于svr_center、svr_login都是http服务器，所以能这么搞
	//如果是tcp服务，就得分成两条消息，让login主动通知client
	isSyncCall := false
	centerSvrId := netConfig.HashCenterID(account)
	netConfig.CallRpcCenter(centerSvrId, enum.Rpc_center_account_login, func(buf *common.NetPack) {
		buf.WriteBuf(req.LeftBuf())
	}, func(recvBuf *common.NetPack) {
		isSyncCall = true
		if err := recvBuf.ReadInt8(); err > 0 {
			//临时读取buffer数据
			oldPos := recvBuf.ReadPos
			accountId := recvBuf.ReadUInt32()
			recvBuf.ReadPos = oldPos

			//生成一个临时token，发给gamesvr、client，用以登录验证
			token := atomic.AddUint32(&g_login_token, 1)
			netConfig.CallRpcGame(gameSvrId, enum.Rpc_game_login_token, func(buf *common.NetPack) {
				buf.WriteUInt32(token)
				buf.WriteUInt32(accountId)
			}, func(recvBuf *common.NetPack) { //返回在线人数
				playerCnt := recvBuf.ReadInt32()
				g_game_player_cnt.Store(gameSvrId, playerCnt)
			})

			ack.WriteInt8(1)
			ack.WriteString(pMeta.OutIP)
			ack.WriteUInt16(pMeta.Port())
			ack.WriteUInt32(token)
			ack.WriteBuf(recvBuf.LeftBuf())
		} else {
			ack.WriteInt8(err)
		}
	})
	if isSyncCall == false {
		panic("Using ack int another CallRpc must be sync!!! zhoumf\n")
	}
}
func Rpc_login_relay_to_center(req, ack *common.NetPack) {
	rpcId := req.ReadUInt16() //目标rpc
	//临时读取buffer数据
	oldPos := req.ReadPos
	strKey := req.ReadString() //accountName/bindVal
	req.ReadPos = oldPos

	svrId := netConfig.HashCenterID(strKey)
	netConfig.SyncRelayToCenter(svrId, rpcId, req, ack)
}

// -------------------------------------
// game svr list
var g_game_player_cnt sync.Map //gameSvrId-playerCnt

func Rpc_login_get_gamesvr_list(req, ack *common.NetPack) {
	version := req.ReadString()

	ids, _ := meta.GetModuleIDs("game", version)
	ack.WriteByte(byte(len(ids)))
	for _, id := range ids {
		pMeta := meta.GetMeta("game", id)
		ack.WriteInt(id)
		ack.WriteString(pMeta.SvrName)
		ack.WriteString(pMeta.OutIP)
		ack.WriteUInt16(pMeta.Port())
		playerCnt := int32(0)
		if v, ok := g_game_player_cnt.Load(id); ok {
			playerCnt = v.(int32)
		}
		ack.WriteInt32(playerCnt)
	}
}
