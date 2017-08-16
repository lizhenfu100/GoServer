package center

import (
	"common"
)

var (
	g_account_login_token = make(map[uint32]uint32)
)

//////////////////////////////////////////////////////////////////////
//!
func Rpc_Login_Token(req, ack *common.NetPack) {
	id := req.ReadUInt32()
	token := req.ReadUInt32()

	g_account_login_token[id] = token
}
func CheckLoginToken(accountId, token uint32) bool {
	if value, ok := g_account_login_token[accountId]; ok {
		return token == value
	}
	return false
}
