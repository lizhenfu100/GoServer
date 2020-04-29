package dbmgo

import (
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

var (
	_actions = make(chan *action, 4096)
	_exit    = make(chan int, 1)
)

const (
	DB_Insert     = 1
	DB_Update     = 2
	DB_Update_Id  = 3
	DB_Update_All = 4
	DB_Remove_One = 5
	DB_Remove_All = 6
	DB_Upsert_Id  = 7
)

type action struct {
	optype byte        //操作类型
	table  string      //表名
	search interface{} //条件
	pData  interface{} //数据，可bson.M指定更新字段，详见dbmgo.go头部注释
}

func WaitStop() {
	if close(_actions); DB() != nil {
		<-_exit
	}
}
func _loop() {
	for v := range _actions {
		if err := v.do(); err != nil {
			gamelog.Error("op[%d] table[%s] search[%v], data[%v], Error[%v]",
				v.optype, v.table, v.search, v.pData, err)
			if isClosedErr(err) { //Mongodb会极低概率忽然断开，所有操作均超时~囧
				for !_default.Connect() {
					time.Sleep(time.Second)
				}
				_actions <- v
			}
		}
	}
	_exit <- 0
}
func (v *action) do() (err error) {
	switch coll := DB().C(v.table); v.optype {
	case DB_Insert:
		err = coll.Insert(v.pData)
	case DB_Update:
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
		if v.search != nil {
			_, err = coll.RemoveAll(v.search)
		}
	}
	return
}
func isClosedErr(e error) bool {
	err := e.Error()
	return strings.LastIndex(err, "timeout") >= 0 ||
		strings.Index(err, "Closed") >= 0
}

func Update(table string, search, update bson.M) {
	_actions <- &action{
		optype: DB_Update,
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
