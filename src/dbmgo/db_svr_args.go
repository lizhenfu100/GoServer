package dbmgo

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	g_svr_args_map sync.Map
)

type kv struct {
	Key string `bson:"_id"`
	Val interface{}
}

func _init_svr_args() {
	var lst []kv
	FindAll("SvrArgs", nil, &lst)
	for _, v := range lst {
		g_svr_args_map.Store(v.Key, v.Val)
	}
}
func GetSvrArg(key string) interface{} {
	if v, ok := g_svr_args_map.Load(key); ok {
		return v
	}
	return nil
}
func SetSvrArg(key string, val interface{}) {
	if _, ok := g_svr_args_map.Load(key); ok {
		UpdateToDB("SvrArgs", bson.M{"_id": key}, bson.M{"$set": bson.M{"val": val}})
	} else {
		InsertToDB("SvrArgs", kv{key, val})
	}
	g_svr_args_map.Store(key, val)
}
