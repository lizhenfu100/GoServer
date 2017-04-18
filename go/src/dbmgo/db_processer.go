package dbmgo

import (
	"conf"
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	g_last_coll  string
	g_db_session *mgo.Session
	g_param_lst  = make(chan *TDB_Param, 1024)
)

type TDB_Param struct {
	isAll  bool    //是否更新全部记录
	coll   string  //集合名
	search *bson.M //条件
	stuff  *bson.M //数据
}

func InitDBProcesser() bool {
	if g_db_session = GetDBSession(); g_db_session == nil {
		gamelog.Error("InitDBProcesser Error : GetDBSession Failed!!!")
		return false
	}
	go _DBProcess()
	return true
}
func _DBProcess() {
	var pColl *mgo.Collection = nil
	var err error
	for param := range g_param_lst {
		if param.coll != g_last_coll {
			pColl = g_db_session.DB(conf.GameDbName).C(param.coll)
			g_last_coll = param.coll
		}
		if param.isAll {
			//nil interface 不等于nil
			if param.search == nil {
				_, err = pColl.UpdateAll(nil, param.stuff)
			} else {
				_, err = pColl.UpdateAll(param.search, param.stuff)
			}
		} else {
			err = pColl.Update(param.search, param.stuff)
		}
		if err != nil {
			gamelog.Error("DBProcess Failed: Collection[%s] search[%v], stuff[%v], Error[%v]",
				param.coll, param.search, param.stuff, err.Error())
		}
	}
}
func UpdateToDB(collection string, search *bson.M, stuff *bson.M) {
	var param TDB_Param
	param.isAll = false
	param.coll = collection
	param.search = search
	param.stuff = stuff
	g_param_lst <- &param
}
func UpdateToDBAll(collection string, search *bson.M, stuff *bson.M) {
	var param TDB_Param
	param.isAll = true
	param.coll = collection
	param.search = search
	param.stuff = stuff
	g_param_lst <- &param
}
