package gamelog

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"
	// _ "go-sql-driver/mysql" //下载并安装mysql驱动
)

// create table logsvr
// (
// 	ID int unsigned not null auto_increment primary key,
// 	EventID int not null,
// 	SrcID int not null,
// 	TargetID int,
// 	Time datetime,
// 	Param1 int,
// 	Param2 int,
// 	Param3 int,
// 	Param4 int
// );
type MSG_SvrLogData struct {
	EventID  int //操作：召唤、购物、打副本等
	SrcID    int //玩家ID
	TargetID int
	Time     int64
	Param    [4]int
}
type TMysqlLog struct {
	db    *sql.DB
	query string
}

const (
	g_user     = "root"
	g_password = ""
	g_addr     = "localhost:3306"
	g_dbname   = "mysql"
)

var (
	g_dsn = fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?timeout=30s&strict=true",
		g_user, g_password, g_addr, g_dbname) //连哪个数据库
)

func NewMysqlLog(table string) *TMysqlLog {
	var err error = nil
	log := new(TMysqlLog)
	log.db, err = sql.Open("mysql", g_dsn)
	if err != nil {
		Error("NewMysqlLog: " + err.Error())
		return nil
	}

	//Notice：sql.Open("mysql", g_dsn) g_dsn为空不会报错，查看源码，只是生成一份数据记录，真正连接数据库是异步的
	//所以这里检查是否真的连上了
	if err = log.db.Ping(); err != nil {
		Error("NewMysqlLog db.Ping(): " + err.Error())
		return nil
	}

	log.query = fmt.Sprintf(
		"INSERT %s SET EventID=?,SrcID=?,TargetID=?,Time=?,Param1=?,Param2=?,Param3=?,Param4=?",
		table) //打开其中哪张表

	return log
}
func (self *TMysqlLog) Close() {
	self.db.Close()
}
func (self *TMysqlLog) Write(data1, data2 [][]byte) {
	//1、开启事务
	tx, err := self.db.Begin()
	if err != nil {
		Error("MysqlLog::db.Begin: " + err.Error())
	}
	stmt, err := tx.Prepare(self.query)
	if err != nil {
		Error("MysqlLog::tx.Prepare: " + err.Error())
	}

	//2、编辑数据
	for _, v := range data1 {
		_transaction(stmt, v)
	}
	for _, v := range data2 {
		_transaction(stmt, v)
	}

	//3、提交事务
	stmt.Close()
	tx.Commit()
}
func _transaction(stmt *sql.Stmt, pdata []byte) {
	// 将buf解析为结构体
	buf := bytes.NewBuffer(pdata)
	dec := gob.NewDecoder(buf)
	var req MSG_SvrLogData
	if dec.Decode(&req) != nil {
		Error("MysqlLog::Transaction : Message Reader!")
		return
	}
	_Exec(stmt, &req)
}
func _Exec(stmt *sql.Stmt, pMsg *MSG_SvrLogData) {
	timeStr := time.Unix(pMsg.Time, 0).Format("2006-01-02 15:04:05")
	_, err := stmt.Exec(pMsg.EventID, pMsg.SrcID, pMsg.TargetID, timeStr, pMsg.Param[0], pMsg.Param[1], pMsg.Param[2], pMsg.Param[3])
	if err != nil {
		Error("MysqlLog::Exec: " + err.Error())
		return
	}
}

func (self *TMysqlLog) InsertDB(pdata *MSG_SvrLogData) {
	// 直接插入数据库
	stmt, err := self.db.Prepare(self.query)
	if err != nil {
		Error("MysqlLog::db.Prepare: " + err.Error())
		return
	}
	defer stmt.Close()

	_Exec(stmt, pdata)
}
