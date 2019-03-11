/***********************************************************************
* @ Mongodb的API
* @ brief
	1、考虑加个异步读接口，传入callback，读到数据后执行
			支持轻量线程的架构里
			是否比“同步读-处理-再写回”的方式好呢？

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
	"common/wechat"
	"fmt"
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	g_db_session *mgo.Session
	g_database   *mgo.Database
)

func InitWithUser(ip string, port uint16, dbname, username, password string) {
	pInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%d", ip, port)},
		Timeout:  10 * time.Second,
		Database: dbname,
		Username: username,
		Password: password,
	}
	var err error
	if g_db_session, err = mgo.DialWithInfo(pInfo); err != nil {
		panic("Mongodb Init Failed:" + err.Error())
	}
	g_database = g_db_session.DB(dbname)

	//FIXME：测试db连接可用性（小概率db初始化成功，但后续读写操作均超时）
	if _, e := Find(kDBIncTable, "_id", "", &nameId{}); e != nil {
		panic("Mongodb Run Failed:" + err.Error())
	}

	go _loop()
	_init_inc_ids()
}
func DataBase() *mgo.Database { return g_database }

//! operation
func InsertSync(table string, pData interface{}) bool {
	coll := g_database.C(table)
	if err := coll.Insert(pData); err != nil {
		gamelog.Error("InsertSync table[%s], data[%v], Error[%s]", table, pData, err.Error())
		wechat.SendMsg("InsertSync: " + err.Error())
		return false
	}
	return true
}
func UpdateIdSync(table string, id, pData interface{}) bool {
	coll := g_database.C(table)
	if err := coll.UpdateId(id, pData); err != nil {
		gamelog.Error("UpdateSync table[%s] id[%v], data[%v], Error[%s]", table, id, pData, err.Error())
		wechat.SendMsg("UpdateIdSync: " + err.Error())
		return false
	}
	return true
}
func RemoveOneSync(table string, search bson.M) bool {
	coll := g_database.C(table)
	if err := coll.Remove(search); err != nil && err != mgo.ErrNotFound {
		gamelog.Error("RemoveOneSync table[%s] search[%v], Error[%s]", table, search, err.Error())
		return false
	}
	return true
}
func RemoveAllSync(table string, search bson.M) bool {
	coll := g_database.C(table)
	if _, err := coll.RemoveAll(search); err != nil && err != mgo.ErrNotFound {
		gamelog.Error("RemoveAllSync table[%s] search[%v], Error[%s]", table, search, err.Error())
		return false
	}
	return true
}
func Find(table, key string, value, pData interface{}) (bool, error) {
	coll := g_database.C(table)
	err := coll.Find(bson.M{key: value}).One(pData)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] key[%s] val[%v]", table, key, value)
			return false, nil
		} else {
			gamelog.Error("Find table[%s] key[%s] val[%v], Error[%s]", table, key, value, err.Error())
			wechat.SendMsg("Find: " + err.Error())
			return false, err
		}
	}
	return true, nil
}
func FindEx(table string, search bson.M, pData interface{}) (bool, error) {
	coll := g_database.C(table)
	if err := coll.Find(search).One(pData); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] search[%v]", table, search)
			return false, nil
		} else {
			gamelog.Error("FindEx table[%s] search[%v], Error[%s]", table, search, err.Error())
			wechat.SendMsg("FindEx: " + err.Error())
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
*/
func FindAll(table string, search bson.M, pSlice interface{}) error {
	coll := g_database.C(table)
	if err := coll.Find(search).All(pSlice); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] search[%v]", table, search)
		} else {
			gamelog.Error("FindAll table[%s] search[%v], Error[%s]", table, search, err.Error())
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
	coll := g_database.C(table)
	query := coll.Find(nil).Sort(sortKey).Limit(cnt)
	if err := query.All(pList); err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Debug("None table[%s] sortKey[%s]", table, sortKey)
		} else {
			gamelog.Error("FindSort table[%s] sortKey[%s] limit[%d], Error[%s]", table, sortKey, cnt, err.Error())
			return err
		}
	}
	return nil
}
