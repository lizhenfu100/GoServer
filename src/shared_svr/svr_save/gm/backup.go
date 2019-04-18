/***********************************************************************
* @ 本地文件备份存档
* @ brief
	、自动运行，按配置定时开启
		· 周五凌晨自动开启，24h后结束……期间所有玩家存档都文件备份
		· 每人每天只备份一次
	、在线人数阀门
		· 超过设定值，关闭此功能（如1w人在线，就只记进度回退的）

	、存档节点是个高宽带、低cpu的服务器，大量的文件写入不大影响正常功能

	、大概的GM接口：
		· 各游戏配置编辑页面
		· 自动运行开关(默认开启)
		· 强制结束
		· 强制开启

* @ author zhoumf
* @ date 2019-3-29
***********************************************************************/
package gm

import (
	"common"
	"common/copy"
	"common/std"
	"conf"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"generate_out/rpc/enum"
	"net/http"
	"netConfig"
	"netConfig/meta"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var G_Backup = &Backup{DBKey: "backup", IsAutoOpen: true}
var g_keyMap sync.Map //<pSave.key, Empty>

type Backup struct {
	sync.RWMutex
	DBKey       string   `bson:"_id"`
	WeekDays    std.Ints //[1, 2]  周几开启
	OnlintLimit int32    //在线量超过，关闭备份
	IsAutoOpen  bool     //是否自动开启，按WeekDays时间

	isOpen int32 //原子读写
}

// ------------------------------------------------------------
// GM接口
func Http_backup_conf(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	G_Backup.Lock()
	copy.CopyForm(G_Backup, q)
	G_Backup.UpdateDB()
	ack, _ := json.MarshalIndent(G_Backup, "", "     ")
	G_Backup.Unlock()
	w.Write(ack)
	gamelog.Info("Http_backup_conf: %v", q)
}
func Http_backup_auto(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	n, _ := strconv.Atoi(q.Get("auto"))
	G_Backup.Lock()
	G_Backup.IsAutoOpen = n > 0
	ack, _ := json.MarshalIndent(G_Backup, "", "     ")
	G_Backup.Unlock()
	w.Write(ack)
	gamelog.Info("Http_backup_auto: %v", q)
}
func Http_backup_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	n, _ := strconv.Atoi(q.Get("force"))
	atomic.StoreInt32(&G_Backup.isOpen, int32(n))
	G_Backup.RLock()
	ack, _ := json.MarshalIndent(G_Backup, "", "     ")
	G_Backup.RUnlock()
	w.Write(ack)
	gamelog.Info("Http_backup_force: %v", q)
}

// ------------------------------------------------------------
func (self *Backup) InitDB() {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.DBKey, self); !ok {
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
	if self.WeekDays.Index(int(time.Now().Weekday())) < 0 {
		atomic.StoreInt32(&G_Backup.isOpen, 0)
	} else if self.IsAutoOpen {
		atomic.StoreInt32(&G_Backup.isOpen, 1)
	}
}
func (self *Backup) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.DBKey, self) }
func (self *Backup) IsValid(key string) bool {
	if atomic.LoadInt32(&G_Backup.isOpen) > 0 {
		if _, ok := g_keyMap.Load(key); !ok {
			g_keyMap.Store(key, std.Empty{})
			return true
		}
	}
	return false
}

// ------------------------------------------------------------
//
func (self *Backup) OnEnterNextDay() {
	G_Backup.RLock()
	idx := G_Backup.WeekDays.Index(int(time.Now().Weekday()))
	isAutoOpen := G_Backup.IsAutoOpen
	G_Backup.RUnlock()
	if idx < 0 {
		atomic.StoreInt32(&G_Backup.isOpen, 0)
	} else if isAutoOpen {
		atomic.StoreInt32(&G_Backup.isOpen, 1)
		g_keyMap = sync.Map{}
	}
}
func (self *Backup) OnEnterNextHour() {
	G_Backup.RLock()
	kOnlintLimit := G_Backup.OnlintLimit
	G_Backup.RUnlock()

	// 查看svr_game在线量，有超过限制值时，关闭备份
	netConfig.CallRpcLogin(enum.Rpc_login_get_game_list, func(buf *common.NetPack) {
		buf.WriteString(meta.G_Local.Version)
	}, func(backBuf *common.NetPack) {
		for cnt, i := backBuf.ReadByte(), byte(0); i < cnt; i++ {
			backBuf.ReadInt()
			backBuf.ReadString()
			backBuf.ReadString()
			backBuf.ReadUInt16()
			onlineCnt := backBuf.ReadInt32()
			if onlineCnt > kOnlintLimit {
				atomic.StoreInt32(&G_Backup.isOpen, 0)
			}
		}
	})
}
