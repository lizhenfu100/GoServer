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
		、解决冲突，记得修改相关状态，在真正写库之前
			· 变更playerId
			· 变更center里的loginSvrId、gameSvrId
		、旧节点DB，过几个月再清理

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
	"shared_svr/svr_save/logic"
	"time"
)

func main() {
	var svrId int
	flag.IntVar(&svrId, "id", 1, "svrId")
	flag.Parse()

	gamelog.InitLogger("MergeSvr")
	InitConf()

	meta.G_Local = meta.GetMeta("save", svrId)

	pMeta := meta.GetMeta("db_save", svrId)
	dbmgo.InitWithUser(pMeta.IP, pMeta.Port(), pMeta.SvrName,
		conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)

	do1()

	fmt.Println("\n...finish...")
	time.Sleep(time.Hour)
}
func InitConf() {
	var metaCfg []meta.Meta
	file.G_Csv_Map = map[string]interface{}{
		"csv/conf_net.csv": &metaCfg,
		"csv/conf_svr.csv": &conf.SvrCsv,
	}
	file.LoadAllCsv()
	meta.InitConf(metaCfg)
	console.Init()
}

// ------------------------------------------------------------
func do1() {
	var list []logic.TSaveData
	dbmgo.FindAll("Save", nil, &list)

	//TODO: 合服工具
	for i := 0; i < len(list); i++ {
		list[i].init()
	}
}
