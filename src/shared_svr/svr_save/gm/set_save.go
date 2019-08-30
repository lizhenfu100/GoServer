package gm

import (
	"common"
	"conf"
	"dbmgo"
	"encoding/json"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"path/filepath"
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
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok && ptr.MacCnt > 0 {
		dbmgo.UpdateId(logic.KDBSave, ptr.Key, bson.M{"$inc": bson.M{"maccnt": -1}})
		w.Write(common.S2B("clear_maccnt: ok"))
	} else {
		w.Write(common.S2B("none save data"))
	}
}
func Http_clear_bind_limit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	key := logic.GetSaveKey(pf_id, uid)
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", key, &logic.TSaveData{}); ok {
		dbmgo.UpdateId(logic.KDBSave, key, bson.M{"$set": bson.M{"chtime": 0}})
		w.Write(common.S2B("clear_bind_limit: ok"))
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

// 存档回退
func Http_show_backup_file(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	key := logic.GetSaveKey(pf_id, uid)

	if names, e := filepath.Glob(fmt.Sprintf("player/%s/*.save", key)); e == nil {
		ack, _ := json.MarshalIndent(names, "", "     ")
		w.Write(ack)
	} else {
		w.Write(common.S2B("none backup data"))
	}
}
func Http_save_backup(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")
	filename := q.Get("filename")

	ack := "unknow error"
	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); !ok {
		ack = "none save data"
	} else if ptr.RollBack(filename) != err.Success {
		ack = "fail to rollback"
	} else {
		ack = "save_rollback: ok"
	}
	w.Write(common.S2B(ack))
}
