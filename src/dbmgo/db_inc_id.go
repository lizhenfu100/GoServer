package dbmgo

import (
	"common"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	g_inc_id_map sync.Map
)

func _init_inc_ids() {
	var lst []common.KeyPair
	FindAll("IncId", nil, &lst)
	for _, v := range lst {
		g_inc_id_map.Store(v.Name, uint32(v.ID))
	}
}
func GetNextIncId(key string) (ret uint32) {
	ret = 1
	if v, ok := g_inc_id_map.Load(key); ok {
		ret += v.(uint32)
	}
	g_inc_id_map.Store(key, ret)
	if ret == 1 {
		InsertToDB("IncId", common.KeyPair{key, 1})
	} else {
		UpdateToDB("IncId", bson.M{"_id": key}, bson.M{"$set": bson.M{"id": ret}})
	}
	return ret
}
