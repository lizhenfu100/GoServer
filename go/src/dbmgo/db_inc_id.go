package dbmgo

import (
	"common"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

var (
	g_inc_id_mutex sync.Mutex
	g_inc_id_map   = make(map[string]uint32)
)

func _init_inc_ids() {
	var lst []common.KeyPair
	FindAll("IncId", nil, &lst)
	for _, v := range lst {
		g_inc_id_map[v.Name] = uint32(v.ID)
	}
}
func GetNextIncId(key string) uint32 {
	g_inc_id_mutex.Lock()
	ret := g_inc_id_map[key] + 1
	g_inc_id_map[key] = ret
	g_inc_id_mutex.Unlock()
	if ret == 1 {
		InsertToDB("IncId", common.KeyPair{key, 1})
	} else {
		UpdateToDB("IncId", bson.M{"_id": key}, bson.M{"$set": bson.M{"id": ret}})
	}
	return ret
}
