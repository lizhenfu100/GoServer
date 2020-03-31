package gm

import (
	"common"
	"common/std/sign/aes"
	"conf"
	"dbmgo"
	"fmt"
	"net/http"
	"shared_svr/svr_sdk/NRIC"
)

func Http_get_account_nric(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	str := q.Get("aid_mac")
	v := NRIC.NRIC{AidMac: NRIC.Parse(str)}
	if ok, _ := dbmgo.Find(NRIC.KDBTable, "_id", v.AidMac, &v); ok {
		w.Write(common.S2B(fmt.Sprintf("%s %s\n生日时间戳：%d\n玩家哈希：%d\n修改次数：%d",
			common.B2S(aes.Decode(v.ID)),
			common.B2S(aes.Decode(v.Name)),
			v.Birthday,
			v.PersonHash,
			v.ChTimes)))
	} else {
		w.Write([]byte("not found"))
	}
}
