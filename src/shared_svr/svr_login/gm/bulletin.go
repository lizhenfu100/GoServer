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
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"reflect"
)

const (
	kDBKey = "bulletin"
	kAllPf = "all_pf" //全平台的公告
)

type Bulletin struct {
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
type Bulletins struct {
	DBKey string `bson:"_id"`
	Pf    map[string]Bulletin
}

func Rpc_login_bulletin(req, ack *common.NetPack) {
	area := req.ReadString()
	pf_id := req.ReadString()

	pAll, ret := &Bulletins{}, ""
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", kDBKey, pAll); ok {
		var pVal *Bulletin
		if v, ok := pAll.Pf[pf_id]; ok { //优先找本平台的
			pVal = &v
		} else if v, ok = pAll.Pf[kAllPf]; !ok { //再找全平台的
			pVal = &v
		}
		if pVal != nil {
			ref := reflect.ValueOf(pVal).Elem()
			if v := ref.FieldByName(area); v.IsValid() {
				ret = v.String()
			}
		}
	}
	ack.WriteString(ret)
}
func Http_bulletin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	pf_id := q.Get("pf_id")
	if pf_id == "" {
		pf_id = kAllPf
	}

	pVal, pAll := &Bulletin{}, &Bulletins{}
	copy.CopyForm(pVal, q)
	pAll.setPlatform(pf_id, pVal)

	ack, _ := json.MarshalIndent(pAll, "", "     ")
	w.Write(ack)
	gamelog.Info("Http_bulletin: %v", q)
}
func Http_view_bulletin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	pf_id := q.Get("pf_id")
	if pf_id == "" {
		pf_id = kAllPf
	}

	pAll := &Bulletins{}
	dbmgo.Find(dbmgo.KTableArgs, "_id", kDBKey, pAll)
	ack, _ := json.MarshalIndent(pAll.Pf[pf_id], "", "     ")
	w.Write(ack)
}

// ------------------------------------------------------------
func (self *Bulletins) setPlatform(pf_id string, pVal *Bulletin) {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", kDBKey, self); ok {
		dbmgo.UpdateId(dbmgo.KTableArgs, kDBKey, bson.M{"$set": bson.M{"pf." + pf_id: pVal}})
	} else {
		self.Pf[pf_id] = *pVal
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
}
