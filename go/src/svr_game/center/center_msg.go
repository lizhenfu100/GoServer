package center

import (
	"common"
	"net/http"
)

var (
	g_account_login_token = make(map[uint32]uint32)
)

//////////////////////////////////////////////////////////////////////
//!
func Handle_Login_Token(w http.ResponseWriter, r *http.Request) {
	//! 接收信息
	req := common.NewByteBufferLen(int(r.ContentLength))
	r.Body.Read(req.DataPtr)

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
