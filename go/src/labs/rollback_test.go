//! 内存数据库事务demo
package main

import (
	"fmt"
	"testing"
)

type TransLog interface {
	Commit(*Database)
	Rollback(*Database)
}
type PlayerItem struct {
	Id     int
	ItemId int
	Num    int
}
type Database struct {
	transLogs  []TransLog
	playerItem map[int]*PlayerItem
}
type TransType int
type PlayerItemTransLog struct {
	Type TransType
	Old  *PlayerItem
	New  *PlayerItem
}

const (
	INSERT TransType = iota
	DELETE
	UPDATE
)

func NewDatabase() *Database {
	return &Database{
		playerItem: make(map[int]*PlayerItem),
	}
}
func (self *Database) Transaction(trans func() int) {
	if trans() < 0 {
		for i := len(self.transLogs) - 1; i >= 0; i-- {
			self.transLogs[i].Rollback(self)
		}
	} else {
		for i := 0; i < len(self.transLogs); i++ {
			self.transLogs[i].Commit(self)
		}
	}
	self.transLogs = self.transLogs[0:0] //每次事务过后清空记录
}

//////////////////////////////////////////////////////////////////////
//! 事务：提交、回退
//////////////////////////////////////////////////////////////////////
func (self *PlayerItemTransLog) Commit(db *Database) {
	switch self.Type {
	case INSERT:
		fmt.Printf(
			"INSERT INTO player_item (id, item_id, num) VALUES (%d, %d, %d)\n",
			self.New.Id, self.New.ItemId, self.New.Num,
		)
	case DELETE:
		fmt.Printf(
			"DELETE player_item WHERE id = %d\n",
			self.Old.Id,
		)
	case UPDATE:
		fmt.Printf(
			"UPDATE player_item SET id = %d, item_id = %d, num = %d\n",
			self.New.Id, self.New.ItemId, self.New.Num,
		)
	}
}
func (self *PlayerItemTransLog) Rollback(db *Database) {
	switch self.Type {
	case INSERT:
		delete(db.playerItem, self.New.Id)
		fmt.Println("Rollback INSERT")
	case DELETE:
		db.playerItem[self.Old.Id] = self.Old
		fmt.Println("Rollback DELETE")
	case UPDATE:
		db.playerItem[self.Old.Id] = self.Old
		fmt.Println("Rollback UPDATE")
	}
}

//////////////////////////////////////////////////////////////////////
//! 数据库操作
//////////////////////////////////////////////////////////////////////
func (self *Database) LookupPlayerItem(id int) *PlayerItem {
	return self.playerItem[id]
}
func (self *Database) InsertPlayerItem(playerItem *PlayerItem) {
	id := playerItem.Id
	self.playerItem[id] = playerItem
	self.transLogs = append(self.transLogs, &PlayerItemTransLog{
		Type: INSERT, New: playerItem,
	})
}
func (self *Database) DeletePlayerItem(id int) {
	old := self.playerItem[id]
	delete(self.playerItem, id)
	self.transLogs = append(self.transLogs, &PlayerItemTransLog{
		Type: DELETE, Old: old,
	})
}
func (self *Database) UpdatePlayerItem(playerItem *PlayerItem) {
	id := playerItem.Id
	old := self.playerItem[id]
	self.playerItem[id] = playerItem
	self.transLogs = append(self.transLogs, &PlayerItemTransLog{
		Type: UPDATE, Old: old, New: playerItem,
	})
}

func Test_main(t *testing.T) {
	db := NewDatabase()

	db.Transaction(func() int {
		db.InsertPlayerItem(&PlayerItem{
			Id:     1,
			ItemId: 100,
			Num:    1,
		})
		db.InsertPlayerItem(&PlayerItem{
			Id:     2,
			ItemId: 100,
			Num:    1,
		})
		return 1
	})
	fmt.Println(db.playerItem, "\n")

	db.Transaction(func() int {
		db.UpdatePlayerItem(&PlayerItem{
			Id:     1,
			ItemId: 111,
			Num:    111,
		})
		return -1
	})
	fmt.Println(*db.playerItem[1], "\n")
}
