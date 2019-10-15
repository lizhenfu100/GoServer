/***********************************************************************
* @ 入库的全局参数
* @ brief
	、本模块的接口性能不高（数据库同步操作）
	、业务模块管理各自的参数缓存

* @ author zhoumf
* @ date 2018-12-7
***********************************************************************/
package dbmgo

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

const KTableArgs = "args"

//type IArgs interface {
//	ReadDB() bool //return Find(dbmgo.KTableArgs, "_id", DBKey, pVal)
//	UpdateDB()    //UpdateId(dbmgo.KTableArgs, DBKey, pVal)
//	InsertDB()    //Insert(dbmgo.KTableArgs, pVal)
//	InitDB()      // if !Find() { Insert() }
//}

// ------------------------------------------------------------
const KTableLog = "log"

type log struct { //多节点取自增id可能重复，导致写入失败
	K1   string
	K2   string
	V    string
	Time int64
}

func Log(key1, key2, val string) {
	Insert(KTableLog, &log{
		key1, key2, val,
		time.Now().Unix(),
	})
}
func LogFind(key1, key2 string) []string {
	if key1 == "" {
		return nil
	}
	var list []log
	FindAll(KTableLog, bson.M{"k2": key2, "k1": key1}, &list)
	ret := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		ret[i] = list[i].V
	}
	return ret
}
