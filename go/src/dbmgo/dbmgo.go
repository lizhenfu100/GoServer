package dbmgo

import (
	"gamelog"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	g_db_session *mgo.Session
	g_database   *mgo.Database
)

func Init(addr, dbname string) {
	var err error
	g_db_session, err = mgo.Dial(addr)
	if err != nil {
		gamelog.Error(err.Error())
		panic("Mongodb Init Failed " + err.Error())
	}
	g_db_session.SetPoolLimit(20)
	g_database = g_db_session.DB(dbname)
	go _DBProcess()
}
func InitWithUser(addr, dbname, username, password string) {
	mgoDialInfo := mgo.DialInfo{
		Addrs:     []string{addr},
		Timeout:   5 * time.Second,
		Username:  username,
		Password:  password,
		PoolLimit: 20,
	}
	var err error
	if g_db_session, err = mgo.DialWithInfo(&mgoDialInfo); err != nil {
		gamelog.Error(err.Error())
		panic("Mongodb Init Failed " + err.Error())
	}
	g_database = g_db_session.DB(dbname)
	go _DBProcess()
}

//! operation
func InsertToDB(table string, pData interface{}) bool {
	coll := g_database.C(table)
	err := coll.Insert(pData)
	if err != nil {
		if !mgo.IsDup(err) {
			gamelog.Error("InsertToDB Failed: table:[%s] Error:[%s]", table, err.Error())
		} else {
			gamelog.Warn("InsertToDB Failed: table:[%s] Error:[%s]", table, err.Error())
		}
		return false
	}
	return true
}
func RemoveFromDB(table string, search *bson.M) error {
	coll := g_database.C(table)
	return coll.Remove(search)
}
func Find(table, key string, value interface{}, pData interface{}) {
	coll := g_database.C(table)
	err := coll.Find(bson.M{key: value}).One(pData)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find table: %s  find: %s:%v", table, key, value)
		} else {
			gamelog.Error3("Find error: %v \r\ntable: %s \r\nfind: %s:%v \r\n",
				err.Error(), table, key, value)
		}
	}
}
func Find_Asc(table, key string, cnt int, pList interface{}) { //升序
	sortKey := "+" + key
	_find_sort(table, sortKey, cnt, pList)
}
func Find_Desc(table, key string, cnt int, pList interface{}) { //降序
	sortKey := "-" + key
	_find_sort(table, sortKey, cnt, pList)
}
func _find_sort(table, sortKey string, cnt int, pList interface{}) {
	coll := g_database.C(table)
	query := coll.Find(nil).Sort(sortKey).Limit(cnt)
	err := query.All(pList)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find")
		} else {
			gamelog.Error3("Find_Sort error: %v \r\ntable: %s \r\nsort: %s \r\nlimit: %d\r\n",
				err.Error(), table, sortKey, cnt)
		}
	}
}
