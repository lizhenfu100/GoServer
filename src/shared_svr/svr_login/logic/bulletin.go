/***********************************************************************
* @ 告示 游戏初始界面的（不登录也会看到）
* @ brief

* @ author zhoumf
* @ date 2019-2-14
***********************************************************************/
package logic

import (
	"common"
	"dbmgo"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"
)

var g_bulletin = Bulletin{DBKey: "bulletin"}

type Bulletin struct {
	sync.Mutex
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

	g_bulletin.Lock()
	ref, ret := reflect.ValueOf(&g_bulletin).Elem(), ""
	if v := ref.FieldByName(area); v.IsValid() {
		ret = v.String()
	}
	g_bulletin.Unlock()
	ack.WriteString(ret)
}
func Http_bulletin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	g_bulletin.Lock()
	common.CopyForm(&g_bulletin, r.Form)
	ack, _ := json.MarshalIndent(&g_bulletin, "", "     ")
	g_bulletin.UpdateDB()
	g_bulletin.Unlock()
	w.Write(ack)
}

// ------------------------------------------------------------
func (self *Bulletin) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.DBKey, self) }
func (self *Bulletin) InitDB() {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.DBKey, self); !ok {
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
}

// ------------------------------------------------------------
func Http_timestamp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%d", time.Now().Unix())))
}
