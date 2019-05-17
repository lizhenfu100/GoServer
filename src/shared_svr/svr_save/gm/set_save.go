package gm

import (
	"common"
	"conf"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"shared_svr/svr_save/logic"
)

func Http_clear_maccnt(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok {
		ptr.MacCnt = 0
		dbmgo.UpdateId(logic.KDBSave, ptr.Key, bson.M{"$set": bson.M{
			"maccnt": ptr.MacCnt}})
		w.Write(common.S2B("clear_maccnt: ok"))
	} else {
		w.Write(common.S2B("none save data"))
	}
}
func Http_clear_unbind_limit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	logic.ClearUnbindLimit()
	w.Write(common.S2B("clear_unbind_limit: ok"))
}

func Http_del_save_data(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok {
		ptr.Backup()
		dbmgo.Remove(logic.KDBSave, bson.M{"_id": ptr.Key})
		dbmgo.RemoveAll(logic.KDBMac, bson.M{"key": ptr.Key})
	}
	w.Write(common.S2B("del_save_data: ok"))
}
