package logic

import (
	"common"
	"encoding/json"
	"fmt"
	"generate_out/err"
	"net/http"
	"shared_svr/svr_center/account"
	"strconv"
)

func Http_show_account(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")
	k := q.Get("k")
	v := q.Get("v")

	var ptr *account.TAccount
	switch k {
	case "account":
		ptr = account.GetAccountByName(v)
	case "aid":
		id, _ := strconv.Atoi(v)
		ptr = account.GetAccountById(uint32(id))
	case "email":
		ptr = account.GetAccountByBindInfo("email", v)
	}
	if ptr != nil {
		//账号
		accountStr, _ := json.MarshalIndent(ptr, "", "     ")
		w.Write(accountStr)
		//所在大区
		buf := common.NewNetPackCap(128)
		ptr.WriteLoginAddr(gameName, buf)
		if e := buf.ReadUInt16(); e == err.Success {
			loginIp := buf.ReadString()
			loginPort := buf.ReadUInt16()
			gameIp := buf.ReadString()
			gamePort := buf.ReadUInt16()
			w.Write(common.S2B(fmt.Sprintf("\nLoginAddr: %s:%d\nGameAddr: %s:%d\nSaveAddr: %s:7090",
				loginIp, loginPort, gameIp, gamePort, gameIp)))
		}
	} else {
		w.Write(common.S2B("none account"))
	}
}
