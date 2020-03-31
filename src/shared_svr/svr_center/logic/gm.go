package logic

import (
	"common"
	"encoding/json"
	"fmt"
	"generate_out/err"
	"net/http"
	"netConfig/meta"
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
	case "aid":
		id, _ := strconv.Atoi(v)
		_, ptr = account.GetAccountById(uint32(id))
	default:
		_, ptr = account.GetAccountByBindInfo(k, v)
	}
	if ptr != nil {
		//账号
		b, _ := json.MarshalIndent(ptr, "", "     ")
		w.Write(b)
		//所在大区
		buf := common.NewNetPackCap(32)
		ptr.WriteLoginAddr(gameName, buf)
		if e := buf.ReadUInt16(); e == err.Success {
			loginIp := buf.ReadString()
			loginPort := buf.ReadUInt16()
			gameIp := buf.ReadString()
			gamePort := buf.ReadUInt16()
			w.Write(common.S2B(fmt.Sprintf("\nLoginAddr: %s:%d\nGameAddr: %s:%d\nSaveAddr: %s:%d",
				loginIp, loginPort, gameIp, gamePort, gameIp, meta.KSavePort)))
		}
		buf.Free()
	} else {
		w.Write([]byte("none account"))
	}
}
