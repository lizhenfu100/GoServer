/***********************************************************************
* @ 玩家登录
* @ brief
	1、验证相关，放登录服，减缓游戏服压力，比如：
		、每个账号在同一区服，只允许建一个或几个角色
		、角色名不能跟别人重复（要查找整张表呀）

	2、登录服验证通过后，将玩家数据载入DBCache(比如redis)，下发AccountId/PlayerId之类的给Client

	3、Client再到游戏服真正登录时，就只需从DBCache载入了

* @ author zhoumf
* @ date 2017-4-26
***********************************************************************/
package player

import (
	"common"
	"generate_out/err"
	"nets/tcp"
	"sync"
	"sync/atomic"
)

func Rpc_check_identity(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	token := req.ReadUInt32()
	if !CheckLoginToken(accountId, token) {
		ack.WriteUInt16(err.Token_verify_err)
	} else if this := FindWithDB(accountId); this == nil {
		ack.WriteUInt16(err.Account_have_none_player)
	} else if this.IsForbidden {
		ack.WriteUInt16(err.Account_forbidden)
	} else {
		this.Login(conn)
		ack.WriteUInt16(err.Success)
		ack.WriteUInt32(this.PlayerID)
		ack.WriteString(this.Name)
	}
}
func Rpc_game_player_set_name(req, ack *common.NetPack, this *TPlayer) {
	this.Name = req.ReadString()
}

// -------------------------------------
// -- 后台账号验证
var g_tokens sync.Map //<accountId, token>

func Rpc_set_identity(req, ack *common.NetPack, conn *tcp.TCPConn) {
	token := req.ReadUInt32()
	accountId := req.ReadUInt32()
	g_tokens.Store(accountId, token)
	cnt := atomic.LoadInt32(&g_online_cnt)
	ack.WriteInt32(cnt + 1)
	//单角色的游戏，自动建号、检查
	ptr := FindWithDB(accountId)
	if ptr == nil {
		ptr = NewPlayerInDB(accountId)
	}
	if ptr == nil {
		ack.WriteUInt16(err.Unknow_error)
	} else if ptr.IsForbidden {
		ack.WriteUInt16(err.Account_forbidden)
	} else {
		ack.WriteUInt16(err.Success)
	}
}
func CheckLoginToken(accountId, token uint32) bool {
	if value, ok := g_tokens.Load(accountId); ok {
		return token == value
	}
	return false
}
