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

const (
	_ = iota
	DB_Insert
	DB_Update_Field
	DB_Update_Id
	DB_Update_All
	DB_Remove_One
	DB_Remove_All
)

type TDB_Param struct {
	optype byte        //操作类型
	table  string      //表名
	search interface{} //条件
	data   interface{} //数据，可bson.M指定更新字段，详见dbmgo.go头部注释
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
		switch param.optype {
		case DB_Insert:
			err = pColl.Insert(param.data)
		case DB_Update_Field:
			err = pColl.Update(param.search, param.data)
		case DB_Update_Id:
			err = pColl.UpdateId(param.search, param.data)
		case DB_Update_All:
			_, err = pColl.UpdateAll(param.search, param.data)
		case DB_Remove_One:
			err = pColl.Remove(param.search)
		case DB_Remove_All:
			_, err = pColl.RemoveAll(param.search)
		}
		if err != nil {
			gamelog.Error("DBProcess Failed: table[%s] search[%v], data[%v], Error[%v]",
				param.table, param.search, param.data, err.Error())
		}
	}
}
func UpdateToDB(table string, search, update bson.M) {
	g_param_chan <- &TDB_Param{
		optype: DB_Update_Field,
		table:  table,
		search: search,
		data:   update,
	}
}
func UpdateIdToDB(table string, id, data interface{}) {
	g_param_chan <- &TDB_Param{
		optype: DB_Update_Id,
		table:  table,
		search: id,
		data:   data,
	}
}
func UpdateAllToDB(table string, search, data bson.M) {
	g_param_chan <- &TDB_Param{
		optype: DB_Update_All,
		table:  table,
		search: search,
		data:   data,
	}
}
func RemoveToDB(table string, search bson.M) {
	g_param_chan <- &TDB_Param{
		optype: DB_Remove_One,
		table:  table,
		search: search,
	}
}
func RemoveAllToDB(table string, search bson.M) {
	g_param_chan <- &TDB_Param{
		optype: DB_Remove_All,
		table:  table,
		search: search,
	}
}
func InsertToDB(table string, data interface{}) {
	g_param_chan <- &TDB_Param{
		optype: DB_Insert,
		table:  table,
		data:   data,
	}
}
