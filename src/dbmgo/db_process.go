package dbmgo

import (
	"gamelog"
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
	DB_Upsert_Id
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
	for v := range _actions {
		var err error
		switch coll := DB().C(v.table); v.optype {
		case DB_Insert:
			err = coll.Insert(v.pData)
		case DB_Update_Field:
			err = coll.Update(v.search, v.pData)
		case DB_Update_Id:
			err = coll.UpdateId(v.search, v.pData)
		case DB_Upsert_Id:
			_, err = coll.UpsertId(v.search, v.pData)
		case DB_Update_All:
			_, err = coll.UpdateAll(v.search, v.pData)
		case DB_Remove_One:
			err = coll.Remove(v.search)
		case DB_Remove_All:
			if v.search == nil {
				err = coll.DropCollection()
			} else {
				_, err = coll.RemoveAll(v.search)
			}
		}
		if err != nil {
			gamelog.Error("DBLoop: op[%d] table[%s] search[%v], data[%v], Error[%v]",
				v.optype, v.table, v.search, v.pData, err)
			if isClosedErr(err) { //Mongodb会极低概率忽然断开，所有操作均超时~囧
				_default.Connect()
				_actions <- v
			}
		}
	}
	_finished.Done()
}
func isClosedErr(e error) bool {
	err := e.Error()
	return strings.LastIndex(err, "timeout") >= 0 ||
		strings.Index(err, "Closed") >= 0
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
func UpsertId(table string, id, pData interface{}) {
	_actions <- &action{
		optype: DB_Upsert_Id,
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
