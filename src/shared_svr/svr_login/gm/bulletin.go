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
	"strconv"
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
		if v, ok := pAll.Pf[pf_id]; ok { //优先找本平台的
			ret = v.Get(area)
		}
		if ret == "" {
			if v, ok := pAll.Pf[kAllPf]; ok { //再找全平台的
				ret = v.Get(area)
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
func (self *Bulletin) Get(area string) string {
	if self != nil {
		ref := reflect.ValueOf(self).Elem()
		if v := ref.FieldByName(area); v.IsValid() {
			return v.String()
		}
	}
	return ""
}

// ------------------------------------------------------------
// 运营用自增id
func Http_get_inc_id(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if key := q.Get("key"); key != "" {
		id := dbmgo.GetNextIncId(key)
		w.Write(common.S2B(strconv.FormatInt(int64(id), 10)))
	}
}
