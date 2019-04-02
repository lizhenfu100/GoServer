/***********************************************************************
* @ 告示 游戏初始界面的（不登录也会看到）
* @ brief

* @ 接口文档
	· Rpc_login_bulletin
	· 上行参数
		· string area		语言地区，须是约定的缩写
	· 下行参数
		· string content	公告文本

* @ author zhoumf
* @ date 2019-2-14
***********************************************************************/
package gm

import (
	"common"
	"common/copy"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"net/http"
	"reflect"
	"sync"
	"time"
)

var g_bulletin = &Bulletin{DBKey: "bulletin"}

type Bulletin struct {
	sync.RWMutex
	DBKey   string `bson:"_id"`
	En      string //告示内容，按国家划分
	Zh      string
	Zh_Hant string
	Jp      string
	Ru      string //俄语
	Kr      string //韩语
	Es      string //西班牙语
	Pt_Br   string //葡萄牙语
	Fr      string //法语
	Id      string //印尼语
	De      string //德语
}

func Rpc_login_bulletin(req, ack *common.NetPack) {
	area := req.ReadString()

	g_bulletin.RLock()
	ref, ret := reflect.ValueOf(g_bulletin).Elem(), ""
	if v := ref.FieldByName(area); v.IsValid() {
		ret = v.String()
	}
	g_bulletin.RUnlock()
	ack.WriteString(ret)
}
func Http_bulletin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	g_bulletin.Lock()
	copy.CopyForm(g_bulletin, r.Form)
	g_bulletin.UpdateDB()
	ack, _ := json.MarshalIndent(g_bulletin, "", "     ")
	g_bulletin.Unlock()
	w.Write(ack)
	gamelog.Info("Http_bulletin: %v", r.Form)
}

// ------------------------------------------------------------
func InitBulletin() {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", g_bulletin.DBKey, g_bulletin); !ok {
		dbmgo.Insert(dbmgo.KTableArgs, g_bulletin)
	}
}
func (self *Bulletin) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.DBKey, self) }

// ------------------------------------------------------------
func Http_timestamp(w http.ResponseWriter, r *http.Request) {
	w.Write(common.S2B(fmt.Sprintf("%d", time.Now().Unix())))
}
