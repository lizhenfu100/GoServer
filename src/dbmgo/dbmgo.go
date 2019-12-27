/***********************************************************************
* @ Mongodb的API
* @ brief
	1、考虑加个异步读接口，传入callback，读到数据后执行
			支持轻量线程的架构里
			是否比“同步读-处理-再写回”的方式好呢？

* @ 研究学习
	· Mongo批量查询，比如查一天内的活跃数据，效率很高，即便time字段没建索引

* @ 几种更新方式
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$set": bson.M{
		"module.data": self.data,
		"goods":    self.Goods,
		"resetday": self.ResetDay}})
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$inc": bson.M{"logincnt": 1}})
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$push": bson.M{"awardlst": pAward}})
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$pushAll": bson.M{"awardlst": awards}})
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$pull": bson.M{
		"bag.items": bson.M{"itemid": itemid}}})
	UpdateToDB("Player", bson.M{"_id": playerID}, bson.M{"$pull": bson.M{
		"bag.items": nil}})

* @ author zhoumf
* @ date 2017-4-22
***********************************************************************/
package dbmgo

import (
	"common/tool/wechat"
	"fmt"
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"sync/atomic"
	"time"
)

var _default DBInfo

type DBInfo struct {
	_pending int32
	sync.RWMutex
	mgo.DialInfo
	db *mgo.Database
}

func InitWithUser(ip string, port uint16, dbname, user, pwd string) {
	_default.Init(ip, port, dbname, user, pwd)
	_init_inc_ids()
	go _loop()
}
func (self *DBInfo) Init(ip string, port uint16, dbname, user, pwd string) {
	self.Addrs = []string{fmt.Sprintf("%s:%d", ip, port)}
	self.Database = dbname
	self.Username = user
	self.Password = pwd
	self.Timeout = 10 * time.Second
	self.Connect()
}
func (self *DBInfo) Connect() {
	if atomic.CompareAndSwapInt32(&self._pending, 0, 1) {
		if session, e := mgo.DialWithInfo(&self.DialInfo); e != nil {
			if self.db == nil {
				panic(fmt.Sprintf("Mongodb Connect: %v", self.DialInfo))
			} else {
				wechat.SendMsg("数据库连接：" + e.Error())
			}
		} else {
			self.Lock()
			self.db = session.DB(self.Database)
			self.Unlock()
		}
		atomic.StoreInt32(&self._pending, 0)
	}
}
func (self *DBInfo) DB() *mgo.Database {
	self.RLock()
	ret := self.db //copy to stack
	self.RUnlock()
	return ret
}
func DB() *mgo.Database { return _default.DB() }

func InsertSync(table string, pData interface{}) bool {
	coll := DB().C(table)
	if err := coll.Insert(pData); err != nil {
		gamelog.Error("InsertSync table[%s] data[%v] Error[%v]", table, pData, err)
		wechat.SendMsg("数据库插入：" + err.Error())
		return false
	}
	return true
}
func UpdateIdSync(table string, id, pData interface{}) bool {
	coll := DB().C(table)
	if err := coll.UpdateId(id, pData); err != nil {
		gamelog.Error("UpdateSync table[%s] id[%v] data[%v] Error[%v]", table, id, pData, err)
		wechat.SendMsg("数据库更新：" + err.Error())
		return false
	}
	return true
}
func UpsertIdSync(table string, id, pData interface{}) bool {
	coll := DB().C(table)
	if _, err := coll.UpsertId(id, pData); err != nil {
		gamelog.Error("UpsertSync table[%s] id[%v] data[%v] Error[%v]", table, id, pData, err)
		wechat.SendMsg("数据库更新：" + err.Error())
		return false
	}
	return true
}
func RemoveOneSync(table string, search bson.M) bool {
	coll := DB().C(table)
	if err := coll.Remove(search); err != nil && err != mgo.ErrNotFound {
		gamelog.Error("RemoveOneSync table[%s] search[%v] Error[%v]", table, search, err)
		return false
	}
	return true
}
func RemoveAllSync(table string, search bson.M) bool {
	coll := DB().C(table)
	if search == nil {
		if e := coll.DropCollection(); e != nil {
			gamelog.Error("Drop table[%s] Error[%v]", table, e)
			return false
		}
	} else if _, err := coll.RemoveAll(search); err != nil && err != mgo.ErrNotFound {
		gamelog.Error("RemoveAllSync table[%s] search[%v] Error[%v]", table, search, err)
		return false
	}
	return true
}

// false && err == nil，才表示db中没有
func Find(table, key string, value, pData interface{}) (bool, error) {
	coll := DB().C(table)
	err := coll.Find(bson.M{key: value}).One(pData)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] key[%s] val[%v]", table, key, value)
			return false, nil
		} else {
			gamelog.Error("Find table[%s] key[%s] val[%v] Error[%v]", table, key, value, err)
			if isClosedErr(err) {
				_default.Connect()
			} else {
				wechat.SendMsg("数据库查询：" + err.Error())
			}
			return false, err
		}
	}
	return true, nil
}
func FindEx(table string, search bson.M, pData interface{}) (bool, error) {
	coll := DB().C(table)
	if err := coll.Find(search).One(pData); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] search[%v]", table, search)
			return false, nil
		} else {
			gamelog.Error("FindEx table[%s] search[%v] Error[%v]", table, search, err)
			if isClosedErr(err) {
				_default.Connect()
			} else {
				wechat.SendMsg("数据库查询：" + err.Error())
			}
			return false, err
		}
	}
	return true, nil
}

/*
=($eq)		bson.M{"name": "Jimmy Kuu"}
!=($ne)		bson.M{"name": bson.M{"$ne": "Jimmy Kuu"}}
>($gt)		bson.M{"age": bson.M{"$gt": 32}}
<($lt)		bson.M{"age": bson.M{"$lt": 32}}
>=($gte)	bson.M{"age": bson.M{"$gte": 33}}
<=($lte)	bson.M{"age": bson.M{"$lte": 31}}
in($in)		bson.M{"name": bson.M{"$in": []string{"Jimmy Kuu", "Tracy Yu"}}}
and			bson.M{"name": "Jimmy Kuu", "age": 33}
or			bson.M{"$or": []bson.M{bson.M{"name": "Jimmy Kuu"}, bson.M{"age": 31}}}
$exists		bson.M{"bindinfo.email": bson.M{ "$exists": false }]
*/
//【学习研究】Mongo批量查询，比如查一天内的活跃数据，效率很高，即便time字段没建索引
func FindAll(table string, search bson.M, pSlice interface{}) error {
	coll := DB().C(table)
	if err := coll.Find(search).All(pSlice); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] search[%v]", table, search)
		} else {
			gamelog.Error("FindAll table[%s] search[%v] Error[%v]", table, search, err)
			return err
		}
	}
	return nil
}
func Find_Asc(table, key string, cnt int, pList interface{}) error { //升序
	sortKey := "+" + key
	return _find_sort(table, sortKey, cnt, pList)
}
func Find_Desc(table, key string, cnt int, pList interface{}) error { //降序
	sortKey := "-" + key
	return _find_sort(table, sortKey, cnt, pList)
}
func _find_sort(table, sortKey string, cnt int, pList interface{}) error {
	coll := DB().C(table)
	query := coll.Find(nil).Sort(sortKey).Limit(cnt)
	if err := query.All(pList); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] sortKey[%s]", table, sortKey)
		} else {
			gamelog.Error("FindSort table[%s] sortKey[%s] limit[%d] Error[%v]", table, sortKey, cnt, err)
			return err
		}
	}
	return nil
}
