package logic

import (
	"common"
	"dbmgo"
	"encoding/json"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"netConfig/meta"
	"shared_svr/svr_center/account"
	"strconv"
	"strings"
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
func Http_game_info(w http.ResponseWriter, r *http.Request) {
	q, p := r.URL.Query(), &account.TAccount{}
	gameName := q.Get("game")
	var ret [][3]int // aid, loginId, gameId
	for _, v := range strings.Split(q.Get("v"), ",") {
		if ok, _ := dbmgo.FindEx(account.KDBTable, bson.M{"$or": []bson.M{
			{"bindinfo.email": v},
			{"bindinfo.phone": v},
			{"bindinfo.name": v},
		}}, p); ok {
			if i, ok := p.GameInfo[gameName]; ok {
				ret = append(ret, [3]int{int(p.AccountID), i.LoginSvrId, i.GameSvrId})
			}
		}
	}
	b, _ := json.Marshal(ret)
	w.Write(b)
}
