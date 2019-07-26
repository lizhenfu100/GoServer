/***********************************************************************
* @ 合服工具，停服期间使用
* @ brief
    1、game、save的合并策略可能不一样
	2、目标节点写入冲突时(如id在目标节点已被占用)，记录旧数据、来源节点……便于出错恢复
	3、合入目标服务器后，需修改center中的游戏登录信息(TGameInfo)

* @ 大家饿
	*、game：无需合并，同大区连的同个db_game
	*、save：
		、依次读取本节点数据库条目，逐个发往目标服，以待入库
		、解决冲突
			· 新存档数据为准
		、变更center里的loginSvrId、gameSvrId
		、旧节点DB，过一周再清理

		、center迁移暂时不搞

* @ author zhoumf
* @ date 2019-3-12
***********************************************************************/
package main

import (
	"common/console"
	"dbmgo"
	"flag"
	"fmt"
	"gamelog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/account"
	"shared_svr/svr_save/logic"
	"time"
)

// ./MergeSvr -addr1 "172.16.0.158" -addr2 "172.16.0.158" -db "save"
func main() {
	var addr1, addr2, dbname string
	flag.StringVar(&addr1, "addr1", "", "addr1")
	flag.StringVar(&addr2, "addr2", "", "addr2")
	flag.StringVar(&dbname, "db", "", "dbname")
	flag.Parse()

	g_dial1.Addrs[0] = addr1+":27017"
	g_dial2.Addrs[0] = addr2+":27017"
	g_dial1.Database = dbname
	g_dial2.Database = dbname

	gamelog.InitLogger("MergeSvr")
	console.Init()

	do1()

	fmt.Println("\n...finish...")
	time.Sleep(time.Hour)
}

// ------------------------------------------------------------
var (
	g_dial1 = &mgo.DialInfo{ //操作节点
		Timeout:  10 * time.Second,
		Addrs:    []string{""},
		Database: "save",
		Username: "chillyroom",
		Password: "db#233*",
	}
	g_dial2 = &mgo.DialInfo{ //目标节点
		Timeout:  10 * time.Second,
		Addrs:    []string{""},
		Database: "save",
		Username: "chillyroom",
		Password: "db#233*",
	}
	g_database1, g_database2 *mgo.Database
)

func do1() {
	dbmgo.Init(g_dial1, &g_database1)
	dbmgo.Init(g_dial2, &g_database2)

	//TODO: 合服工具
	/*
		1、把各个大区的子db合到一块，一个大区一个db

		2、center里的 HappyDiner gameInfo.GameSvrId均改为1  -->
		备份子节点db --> 子节点读数据 --> 发至登录服的svr_save --> 写入db

		3、center迁移至新机器
	*/

	//merge()
	//resetCenterGameInfo()
	//delGameInfo()
	cutSaveMacCnt()
}

func merge() { //读数据，写入主节点
	p1, p2 := &logic.TSaveData{}, &logic.TSaveData{}
	iter1 := g_database1.C(logic.KDBSave).Find(nil).Iter()
	coll2 := g_database2.C(logic.KDBSave)
	for {
		if !iter1.Next(p1) {
			break
		}
		if p1.MacCnt = 0; coll2.Insert(p1) != nil {
			if coll2.Find(bson.M{"_id": p1.Key}).One(p2) != nil && p1.UpTime > p2.UpTime {
				if coll2.UpdateId(p1.Key, p1) != nil {
					gamelog.Error("insert fail: %v", p1)
				}
			}
		}
	}
}
func resetCenterGameInfo() {
	p := &account.TAccount{}
	coll := g_database1.C(account.KDBTable)
	iter := coll.Find(nil).Iter()
	for {
		if !iter.Next(p) {
			break
		}
		if _, ok := p.BindInfo["email"]; !ok {
			if coll.Find(bson.M{"bindinfo.email": p.Name}).One(&account.TAccount{}) == mgo.ErrNotFound {
				p.BindInfo["email"] = p.Name
				coll.UpdateId(p.AccountID, bson.M{"$set": bson.M{
					fmt.Sprintf("bindinfo"): p.BindInfo}})
			}
		}
	}
}
func delGameInfo() {
	p := &account.TAccount{}
	coll := g_database1.C(account.KDBTable)
	iter := coll.Find(nil).Iter()
	for {
		if !iter.Next(p) {
			break
		}
		if v, ok := p.GameInfo["HappyDiner"]; ok && v.LoginSvrId == 0 {
			delete(p.GameInfo, "HappyDiner")
			coll.UpdateId(p.AccountID, bson.M{"$set": bson.M{
				fmt.Sprintf("gameinfo"): p.GameInfo}})
		}
	}
}
func moveCenterDB() {
	p1 := &account.TAccount{}
	iter := g_database1.C(account.KDBTable).Find(nil).Iter()
	coll2 := g_database2.C(account.KDBTable)
	for {
		if !iter.Next(p1) {
			break
		}
		if coll2.Insert(p1) != nil {
			gamelog.Error("insert fail: %v", p1)
		}
	}
}
func cutSaveMacCnt() {
	p := &logic.TSaveData{}
	coll := g_database1.C(logic.KDBSave)
	iter := coll.Find(nil).Iter()
	for {
		if !iter.Next(p) {
			break
		}
		if p.MacCnt >= 2 {
			p.MacCnt -= 2
		} else {
			p.MacCnt = 0
		}
		coll.UpdateId(p.Key, bson.M{"$set": bson.M{"maccnt": p.MacCnt}})
	}
}