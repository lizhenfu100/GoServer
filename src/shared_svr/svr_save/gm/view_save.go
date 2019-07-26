package gm

import (
	"common"
	"conf"
	"dbmgo"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"shared_svr/svr_save/logic"
)

func Http_view_save_data(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok {
		ptr.Data = nil
		ack, _ := json.MarshalIndent(ptr, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B("none save data"))
	}
}

func Http_view_bind_mac(w http.ResponseWriter, r *http.Request) { //绑定的设备信息
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	var list []logic.MacInfo
	key := logic.GetSaveKey(pf_id, uid)
	if e := dbmgo.FindAll(logic.KDBMac, bson.M{"key": key}, &list); e == nil {
		ack, _ := json.MarshalIndent(list, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B("none mac data"))
	}
}

// ------------------------------------------------------------
// 设备码绑定过的所有账户
func Http_find_aid_in_mac(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	mac := q.Get("mac")

	ret := dbmgo.LogFind("UnbindMac", mac)
	pMac := &logic.MacInfo{}
	if ok, _ := dbmgo.Find(logic.KDBMac, "_id", mac, pMac); ok {
		ret = append(ret, pMac.Key)
	}

	ack, _ := json.MarshalIndent(ret, "", "     ")
	w.Write(ack)
}
