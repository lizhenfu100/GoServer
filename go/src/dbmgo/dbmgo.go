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
func InsertSync(table string, pData interface{}) error {
	coll := g_database.C(table)
	return coll.Insert(pData)
}
func RemoveSync(table string, search bson.M) error {
	coll := g_database.C(table)
	return coll.Remove(search)
}
func Find(table, key string, value, pData interface{}) bool {
	coll := g_database.C(table)
	err := coll.Find(bson.M{key: value}).One(pData)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find table: %s  find: %s:%v", table, key, value)
		} else {
			gamelog.Error3("Find error: %v \r\ntable: %s \r\nfind: %s:%v \r\n",
				err.Error(), table, key, value)
		}
		return false
	}
	return true
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
func FindAll(table string, search bson.M, pSlice interface{}) {
	coll := g_database.C(table)
	err := coll.Find(search).All(pSlice)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find table: %s  findall: %v", table, search)
		} else {
			gamelog.Error3("FindAll error: %v \r\ntable: %s \r\nfindall: %v \r\n",
				err.Error(), table, search)
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
