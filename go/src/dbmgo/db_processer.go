package dbmgo

import (
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	g_last_table string
	g_param_chan = make(chan *TDB_Param, 1024)
	g_cache_coll = make(map[string]*mgo.Collection)
)

type TDB_Param struct {
	isAll    bool //是否更新全部记录
	isInsert bool
	table    string //表名
	search   bson.M //条件
	stuff    bson.M //数据
	pData    interface{}
}

func _DBProcess() {
	var pColl *mgo.Collection = nil
	var err error
	var ok bool
	for param := range g_param_chan {
		if param.table != g_last_table {
			if pColl, ok = g_cache_coll[param.table]; !ok {
				pColl = g_database.C(param.table)
				g_cache_coll[param.table] = pColl
			}
			g_last_table = param.table
		}
		if param.isInsert {
			err = pColl.Insert(param.pData)
		} else if param.isAll {
			_, err = pColl.UpdateAll(param.search, param.stuff)
		} else {
			err = pColl.Update(param.search, param.stuff)
		}
		if err != nil {
			gamelog.Error("DBProcess Failed: table[%s] search[%v], stuff[%v], Error[%v]",
				param.table, param.search, param.stuff, err.Error())
		}
	}
}
func UpdateToDB(table string, search, stuff bson.M) {
	var param TDB_Param
	param.isAll = false
	param.table = table
	param.search = search
	param.stuff = stuff
	g_param_chan <- &param
}
func UpdateToDBAll(table string, search, stuff bson.M) {
	var param TDB_Param
	param.isAll = true
	param.table = table
	param.search = search
	param.stuff = stuff
	g_param_chan <- &param
}
func InsertToDB(table string, pData interface{}) {
	var param TDB_Param
	param.isInsert = true
	param.table = table
	param.pData = pData
	g_param_chan <- &param
}
