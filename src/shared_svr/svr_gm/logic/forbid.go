package logic

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

type Info struct {
	Key     string `bson:"_id"`
	Time    int64
	Day     uint16 //封多少天，0永久
	IsWhite bool
}

func Http_names_add(w http.ResponseWriter, r *http.Request) { //GM 加黑、白名单
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	table := q.Get("table") //哪张表
	key := q.Get("key")
	day, _ := strconv.Atoi(q.Get("day"))
	white, _ := strconv.Atoi(q.Get("white"))
	isWhite := white > 0
	timenow := time.Now().Unix()

	ptr := &Info{key, timenow, uint16(day), isWhite}
	if ok, _ := dbmgo.Find(table, "_id", key, ptr); ok {
		if ptr.IsWhite != isWhite {
			ptr.IsWhite = isWhite
			dbmgo.UpdateId(table, key, ptr)
		}
	} else {
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

	ptr := &Info{}
	if ok, _ := dbmgo.Find(table, "_id", key, ptr); ok && !ptr.IsWhite {
		timenow := time.Now().Unix()
		if ptr.Day == 0 || (timenow-ptr.Time)/(3600*24) < int64(ptr.Day) {
			ack.WriteByte(1)
			ack.WriteInt64(ptr.Time)
			ack.WriteUInt16(ptr.Day)
			return
		} else {
			dbmgo.Remove(table, bson.M{"_id": key}) //时间到，解封
		}
	}
	ack.WriteByte(0)
}
func Rpc_gm_forbid_add(req, ack *common.NetPack) {
	table := req.ReadString()
	key := req.ReadString()
	day := req.ReadUInt16()

	timenow := time.Now().Unix()
	dbmgo.UpsertId(table, key, Info{Key: key, Time: timenow, Day: day})
}
func Rpc_gm_forbid_del(req, ack *common.NetPack) {
	table := req.ReadString()
	key := req.ReadString()
	dbmgo.Remove(table, bson.M{"_id": key})
}
