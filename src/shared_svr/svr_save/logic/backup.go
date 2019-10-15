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
package logic

import (
	"common"
	"common/copy"
	"common/file"
	"common/std"
	"conf"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"generate_out/err"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var G_Backup = &Backup{DBKey: "backup"}
var g_keyMap sync.Map //<pSave.key, Empty>

type Backup struct {
	sync.RWMutex `bson:"-"`
	DBKey        string   `bson:"_id"`
	WeekDays     std.Ints //[1, 2]  周几开启
	isOpen       int32    //原子读写
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
	idx := self.WeekDays.Index(int(time.Now().Weekday()))
	atomic.StoreInt32(&self.isOpen, int32(idx+1))
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
func (self *Backup) OnEnterNextDay() {
	G_Backup.RLock()
	idx := G_Backup.WeekDays.Index(int(time.Now().Weekday()))
	G_Backup.RUnlock()
	atomic.StoreInt32(&G_Backup.isOpen, int32(idx+1))
	g_keyMap = sync.Map{} //清空昨日记录
}

// ------------------------------------------------------------
// 敏感数据（如游戏进度）异动，记录历史存档
type TSensitive struct {
	GameSession int //进度，不同游戏含义不一
}

func (self *TSaveData) CheckBackup(newExtra string) {
	pNew, pOld := &TSensitive{}, &TSensitive{}
	json.Unmarshal(common.S2B(newExtra), pNew)
	json.Unmarshal(common.S2B(self.Extra), pOld)
	if pNew.GameSession < pOld.GameSession {
		gamelog.Warn("GameSession rollback: %s", self.Key)
	}
	if pNew.GameSession < pOld.GameSession || G_Backup.IsValid(self.Key) {
		self.Backup()
	}
}
func (self *TSaveData) Backup() {
	dir := fmt.Sprintf("player/%s/", self.Key)
	name := time.Now().Format("20060102_150405") + ".save"
	if fi, e := file.CreateFile(dir, name, os.O_TRUNC|os.O_WRONLY); e == nil {
		if _, e = fi.Write(self.Data); e != nil {
			gamelog.Error("Backup: %s", e.Error())
		}
		fi.Close()
		file.DelExpired(dir, "", 30) //删除30天前的记录
	}
}
func (self *TSaveData) RollBack(filename string) uint16 {
	if f, e := os.Open(fmt.Sprintf("player/%s/%s", self.Key, filename)); e != nil {
		return err.Not_found
	} else {
		if buf, e := ioutil.ReadAll(f); e == nil {
			self.Data = buf
			self.UpTime = time.Now().Unix()
			dbmgo.UpdateIdSync(KDBSave, self.Key, self)
		}
		f.Close()
		return err.Success
	}
}
