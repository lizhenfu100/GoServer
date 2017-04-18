package dbmgo

import (
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "sync"
	"time"
)

var (
	g_db_conn *mgo.Session
	// UpdateLock sync.Mutex
	// InsertLock sync.Mutex
)

func Init(addr string) {
	var err error
	g_db_conn, err = mgo.Dial(addr)
	if err != nil {
		gamelog.Error(err.Error())
		panic("Mongodb Init Failed " + err.Error())
	}
	g_db_conn.SetPoolLimit(20)
}
func InitWithUser(addr string, username string, password string) {
	mgoDialInfo := mgo.DialInfo{
		Addrs:     []string{addr},
		Timeout:   5 * time.Second,
		Username:  username,
		Password:  password,
		PoolLimit: 20,
	}
	var err error
	g_db_conn, err = mgo.DialWithInfo(&mgoDialInfo)
	if err != nil {
		gamelog.Error(err.Error())
		panic("Mongodb Init Failed " + err.Error())
	}
}
func GetDBSession() *mgo.Session {
	if g_db_conn == nil {
		gamelog.Error("GetDBSession Failed, g_db_conn is nil!!")
		panic("db connections is nil!!")
	}
	return g_db_conn.Clone()
}
func InsertToDB(dbname string, collection string, data interface{}) bool {
	s := GetDBSession()
	defer s.Close()
	coll := s.DB(dbname).C(collection)
	err := coll.Insert(&data)
	if err != nil {
		if !mgo.IsDup(err) {
			gamelog.Error("InsertToDB Failed: DB:[%s] Collection:[%s] Error:[%s]", dbname, collection, err.Error())
		} else {
			gamelog.Warn("InsertToDB Failed: DB:[%s] Collection:[%s] Error:[%s]", dbname, collection, err.Error())
		}
		return false
	}
	return true
}
func RemoveFromDB(dbname string, collection string, search *bson.M) error {
	s := GetDBSession()
	defer s.Close()

	coll := s.DB(dbname).C(collection)

	return coll.Remove(search)
}
func Find(dbName string, tableName string, find string, find_value interface{}, data interface{}) int {
	s := GetDBSession()
	defer s.Close()
	collection := s.DB(dbName).C(tableName)
	err := collection.Find(bson.M{find: find_value}).One(data)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find dbName: %s  ntable: %s  find: %s:%v", dbName, tableName, find, find_value)
			return 1
		}
		gamelog.Error3("Find error: %v \r\ndbName: %s \r\ntable: %s \r\nfind: %s:%v \r\n",
			err.Error(), dbName, tableName, find, find_value)
		return -1
	}
	return 0
}

//! 排序查找
//! order 1 -> 正序  -1 -> 倒序
func Find_Sort(dbName string, tableName string, find string, order int, number int, lst interface{}) int {
	s := GetDBSession()
	defer s.Close()

	strSort := ""
	if order == 1 {
		strSort = "+" + find
	} else {
		strSort = "-" + find
	}
	collection := s.DB(dbName).C(tableName)
	query := collection.Find(nil).Sort(strSort).Limit(number)
	err := query.All(lst)
	if err != nil {
		if err == mgo.ErrNotFound {
			gamelog.Warn("Not Find")
			return 1
		}
		gamelog.Error3("Find_Sort error: %v \r\ndbName: %s \r\ntable: %s \r\nfind: %s \r\norder: %d\r\nlimit: %d\r\n",
			err.Error(), dbName, tableName, find, order, number)
		return -1
	}
	return 0
}

/*
func UpdateToDBAll(dbname string, collection string, search *bson.M, stuff *bson.M) bool {
	s := GetDBSession()
	defer s.Close()
	coll := s.DB(dbname).C(collection)
	_, err := coll.UpdateAll(search, stuff)
	if err != nil {
		gamelog.Error3("UpdateToDB Failed: DB:[%s] Collection:[%s] search:[%v], stuff:[%v], Error:%v",
			dbname, collection, search, stuff, err.Error())
		return false
	}
	return true
}
func UpdateToDB(dbname string, collection string, search *bson.M, stuff *bson.M) bool {
	s := GetDBSession()
	defer s.Close()
	coll := s.DB(dbname).C(collection)
	//UpdateLock.Lock()
	//t1 := time.Now().UnixNano()
	err := coll.Update(search, stuff)
	//t2 := time.Now().UnixNano()
	//UpdateLock.Unlock()
	//gamelog.Error("UpdateToDB time:%d", t2-t1)
	if err != nil {
		gamelog.Error3("UpdateToDB Failed: DB:[%s] Collection:[%s] search:[%v], stuff:[%v], Error:%v",
			dbname, collection, search, stuff, err.Error())
		return false
	}
	return true
}
*/
