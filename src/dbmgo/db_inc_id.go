package dbmgo

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

const (
	kDBIncTable = "IncId"
	kUseCache   = false //FIXME：若多个game连同个db，g_inc_id_map的缓存就与db中的不一致了
)

var (
	g_inc_id_map   = make(map[string]uint32)
	g_inc_id_mutex sync.Mutex
)

type nameId struct {
	Name string `bson:"_id"`
	ID   uint32
}

func _init_inc_ids() {
	if kUseCache {
		var lst []nameId
		FindAll(kDBIncTable, nil, &lst)
		g_inc_id_mutex.Lock()
		for _, v := range lst {
			g_inc_id_map[v.Name] = v.ID
		}
		g_inc_id_mutex.Unlock()
	}
}

//返回值不一定都是健壮的，可能有重复（见FIXME注释）
func GetNextIncId(key string) uint32 {
	if kUseCache {
		//实际包含三步（读出、+1、写入）须原子的完成，才可保证每次返回不同id；sync.Map仅保障了读写安全性
		g_inc_id_mutex.Lock()
		ret := g_inc_id_map[key] + 1
		g_inc_id_map[key] = ret
		g_inc_id_mutex.Unlock()
		if ret == 1 {
			Insert(kDBIncTable, &nameId{key, 1})
		} else {
			UpdateId(kDBIncTable, key, bson.M{"$inc": bson.M{"id": 1}}) //$set在多进程架构中有脏写风险
		}
		return ret
	} else {
		v := &nameId{key, 0}
		g_inc_id_mutex.Lock()
		//FIXME：两进程同时读，返回同样结果，其中一个后续操作会失败
		//tcp连同个dbproxy，单线程取
		if ok, _ := Find(kDBIncTable, "_id", key, v); ok {
			v.ID++
			UpdateIdSync(kDBIncTable, key, bson.M{"$inc": bson.M{"id": 1}})
		} else {
			v.ID++
			InsertSync(kDBIncTable, v)
		}
		g_inc_id_mutex.Unlock()
		return v.ID
	}
}
