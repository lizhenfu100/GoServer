/***********************************************************************
* @ 公告（不登录也会看到）
* @ brief
	· 公告可能被所有玩家访问，只读，有分流需求，不适合放单节点模块

* @ 接口文档
	· Rpc_login_bulletin
	· 上行参数
		· string language	语言地区，须是约定的缩写
	· 下行参数
		· string content	公告文本

* @ author zhoumf
* @ date 2019-2-14
***********************************************************************/
package gm

import (
	"common"
	"common/timer"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

const (
	kDBKey = "bulletin"
	kAllPf = "all_pf" //全平台的公告
)

type Bulletin struct {
	Begin int64
	End   int64
	Txt   map[string]string //<language, txt>
	//En   Zh   Zh_Hant Jp Ru Kr  Es   Pt_Br  Fr   Id  De    Ar    Fa
	//英语 简中    繁中	日 俄 韩 西班牙 葡萄牙 法语 印尼 德语 阿拉伯 波斯语
}
type Bulletins struct {
	DBKey string `bson:"_id"`
	Pf    map[string]Bulletin
}

func Rpc_login_bulletin(req, ack *common.NetPack, _ common.Conn) {
	language := req.ReadString()
	pf_id := req.ReadString()
	pAll, ret, timenow := &Bulletins{}, "", time.Now().Unix()
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", kDBKey, pAll); ok {
		if v, ok := pAll.Pf[pf_id]; ok { //优先找本平台的
			if common.InTime(timenow, v.Begin, v.End) {
				ret = v.Txt[language]
			}
		}
		if ret == "" {
			if v, ok := pAll.Pf[kAllPf]; ok { //再找全平台的
				if common.InTime(timenow, v.Begin, v.End) {
					ret = v.Txt[language]
				}
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
	pVal := &Bulletin{Txt: map[string]string{}}
	pVal.Begin = timer.S2T(q.Get("begin"))
	pVal.End = timer.S2T(q.Get("end"))
	delete(q, "pf_id")
	delete(q, "begin")
	delete(q, "end")
	for k, v := range q {
		pVal.Txt[k] = v[0]
	}
	pAll := &Bulletins{}
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
		self.Pf[pf_id] = *pVal
		dbmgo.UpdateId(dbmgo.KTableArgs, kDBKey, bson.M{"$set": bson.M{"pf." + pf_id: pVal}})
	} else {
		self.DBKey = kDBKey
		self.Pf = map[string]Bulletin{pf_id: *pVal}
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
}

// ------------------------------------------------------------
// 运营用自增id，客户端邮件id
func Http_get_inc_id(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if key := q.Get("key"); key != "" {
		id := dbmgo.GetNextIncId(key)
		w.Write(common.S2B(strconv.FormatInt(int64(id), 10)))
	}
}