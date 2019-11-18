package logic

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
)

func Http_names_add(w http.ResponseWriter, r *http.Request) { //GM 加黑、白名单
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	table := q.Get("table") //哪张表
	key := q.Get("key")
	white, _ := strconv.Atoi(q.Get("white"))
	isWhite := white > 0

	ptr := &Msg{Key: key}
	if ok, _ := dbmgo.Find(table, "_id", key, ptr); ok {
		if ptr.IsWhite != isWhite {
			ptr.IsWhite = isWhite
			dbmgo.UpdateId(table, key, ptr)
		}
	} else {
		ptr.IsWhite = isWhite
		dbmgo.Insert(table, ptr)
	}
	gamelog.Info("names_add: %s", r.URL.String())
}
func Http_names_del(w http.ResponseWriter, r *http.Request) { //GM 删黑、白名单
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	table := q.Get("table")
	key := q.Get("key")

	dbmgo.Remove(table, bson.M{"_id": key})
	gamelog.Info("names_del: %v", r.URL.String())
}

func Rpc_gm_forbid_check(req, ack *common.NetPack) {
	table := req.ReadString() //哪张表
	key := req.ReadString()

	ptr := &Msg{}
	if ok, _ := dbmgo.Find(table, "_id", key, ptr); ok && !ptr.IsWhite {
		ack.WriteByte(1)
	} else {
		ack.WriteByte(0)
	}
}
func Rpc_gm_forbid_add(req, ack *common.NetPack) {
	table := req.ReadString()
	key := req.ReadString()

	ptr := &Msg{Key: key}
	if ok, _ := dbmgo.Find(table, "_id", key, ptr); !ok {
		dbmgo.Insert(table, ptr)
	}
}
func Rpc_gm_forbid_del(req, ack *common.NetPack) {
	table := req.ReadString()
	key := req.ReadString()
	dbmgo.Remove(table, bson.M{"_id": key})
}

// ------------------------------------------------------------
type Msg struct {
	Key     string `bson:"_id"`
	IsWhite bool
}
