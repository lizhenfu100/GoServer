package dbmgo

import (
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sync"
)

var (
	_actions  = make(chan *action, 4096)
	_finished sync.WaitGroup //告知DBProcess结束
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

func WaitStop() { close(_actions); _finished.Wait() }

func _loop() {
	_finished.Add(1)
	var err error
	var lastTable string
	var coll *mgo.Collection
	for v := range _actions {
		if v.table != lastTable {
			coll = DB().C(v.table)
			lastTable = v.table
		}
		switch v.optype {
		case DB_Insert:
			err = coll.Insert(v.pData)
		case DB_Update_Field:
			err = coll.Update(v.search, v.pData)
		case DB_Update_Id:
			err = coll.UpdateId(v.search, v.pData)
		case DB_Update_All:
			_, err = coll.UpdateAll(v.search, v.pData)
		case DB_Remove_One:
			err = coll.Remove(v.search)
		case DB_Remove_All:
			_, err = coll.RemoveAll(v.search)
		}
		if err != nil {
			errTips := err.Error()
			gamelog.Error("DBLoop: op[%d] table[%s] search[%v], data[%v], Error[%s]",
				v.optype, v.table, v.search, v.pData, errTips)
			//FIXME：Mongodb会极低概率忽然断开，所有操作均超时~囧
			if strings.LastIndex(errTips, "timeout") >= 0 {
				_default.Connect()
				coll = DB().C(v.table) //重连后缓存须更新
				_actions <- v
			}
		}
	}
	_finished.Done()
}

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
