package dbmgo

import (
	"common/wechat"
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	_actions    = make(chan *action, 4096)
	_finished   = make(chan bool) //告知DBProcess结束
	_last_table string
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

type action struct {
	optype byte        //操作类型
	table  string      //表名
	search interface{} //条件
	pData  interface{} //数据，可bson.M指定更新字段，详见dbmgo.go头部注释
}

func _loop() {
	var err error
	var pColl *mgo.Collection
	for v := range _actions {
		if v.table != _last_table {
			pColl = g_database.C(v.table)
			_last_table = v.table
		}
		switch v.optype {
		case DB_Insert:
			err = pColl.Insert(v.pData)
		case DB_Update_Field:
			err = pColl.Update(v.search, v.pData)
		case DB_Update_Id:
			err = pColl.UpdateId(v.search, v.pData)
		case DB_Update_All:
			_, err = pColl.UpdateAll(v.search, v.pData)
		case DB_Remove_One:
			err = pColl.Remove(v.search)
		case DB_Remove_All:
			_, err = pColl.RemoveAll(v.search)
		}
		if err != nil {
			gamelog.Error("DBProcess Failed: op[%d] table[%s] search[%v], data[%v], Error[%s]",
				v.optype, v.table, v.search, v.pData, err.Error())
			wechat.SendMsg("DBProcess: " + err.Error())
		}
	}
	_finished <- true
}
func WaitStop() { close(_actions); <-_finished }

func Update(table string, search, update bson.M) {
	_actions <- &action{
		optype: DB_Update_Field,
		table:  table,
		search: search,
		pData:  update,
	}
}
func UpdateId(table string, id, pData interface{}) {
	_actions <- &action{
		optype: DB_Update_Id,
		table:  table,
		search: id,
		pData:  pData,
	}
}
func UpdateAll(table string, search, data bson.M) {
	_actions <- &action{
		optype: DB_Update_All,
		table:  table,
		search: search,
		pData:  data,
	}
}
func Remove(table string, search bson.M) {
	_actions <- &action{
		optype: DB_Remove_One,
		table:  table,
		search: search,
	}
}
func RemoveAll(table string, search bson.M) {
	_actions <- &action{
		optype: DB_Remove_All,
		table:  table,
		search: search,
	}
}
func Insert(table string, pData interface{}) {
	_actions <- &action{
		optype: DB_Insert,
		table:  table,
		pData:  pData,
	}
}
