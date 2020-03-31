package dbmgo

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

const kDBIncTable = "IncId"

type nameId struct {
	Name string `bson:"_id"`
	ID   uint32
}

var _inc_mutex sync.Mutex

func GetNextIncId(key string) uint32 {
	v := &nameId{key, 0}
	_inc_mutex.Lock()
	//FIXME：两进程同时读，返回同样结果，其中一个后续操作会失败
	if ok, e := Find(kDBIncTable, "_id", key, v); ok {
		v.ID++
		UpdateIdSync(kDBIncTable, key, bson.M{"$inc": bson.M{"id": 1}})
	} else if e == nil {
		v.ID++
		InsertSync(kDBIncTable, v)
	}
	_inc_mutex.Unlock()
	return v.ID
}
