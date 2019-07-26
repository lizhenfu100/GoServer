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
	"gamelog"
	"net/http"
	"reflect"
)

const kDBKey = "bulletin"

type Bulletin struct {
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

	ptr, ret := &Bulletin{DBKey: kDBKey}, ""
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", ptr.DBKey, ptr); ok {
		ref := reflect.ValueOf(ptr).Elem()
		if v := ref.FieldByName(area); v.IsValid() {
			ret = v.String()
		}
	}
	ack.WriteString(ret)
}
func Http_bulletin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ptr := &Bulletin{DBKey: kDBKey}
	copy.CopyForm(ptr, r.Form)
	dbmgo.UpdateId(dbmgo.KTableArgs, ptr.DBKey, ptr)
	ack, _ := json.MarshalIndent(ptr, "", "     ")
	w.Write(ack)
	gamelog.Info("Http_bulletin: %v", r.Form)
}

// ------------------------------------------------------------
