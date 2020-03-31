package gm

import (
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
		w.Write([]byte("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok {
		ack, _ := json.MarshalIndent(ptr, "", "     ")
		w.Write(ack)
	} else {
		w.Write([]byte("none save data"))
	}
}

func Http_view_bind_mac(w http.ResponseWriter, r *http.Request) { //绑定的设备信息
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	var list []logic.MacInfo
	if dbmgo.FindAll(logic.KDBMac, bson.M{"key": logic.GetSaveKey(pf_id, uid)}, &list) == nil {
		ack, _ := json.MarshalIndent(list, "", "     ")
		w.Write(ack)
	} else {
		w.Write([]byte("none mac data"))
	}
}

// ------------------------------------------------------------
// 设备码绑定过的所有账户
func Http_find_aid_in_mac(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	mac := q.Get("mac")

	ret := []string{"正在使用的："}
	pMac := &logic.MacInfo{}
	if ok, _ := dbmgo.Find(logic.KDBMac, "_id", mac, pMac); ok {
		ret = append(ret, pMac.Key)
	}
	ret = append(ret, "解绑过的：")
	ret = append(ret, dbmgo.LogFind("UnbindMac", mac)...)
	ack, _ := json.MarshalIndent(ret, "", "     ")
	w.Write(ack)
}
