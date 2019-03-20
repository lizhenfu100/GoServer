package account

import (
	"common"
	"conf"
	"dbmgo"
	"gamelog"
	"net/http"
	"sync"
)

var G_WhiteList = WhiteList{DBKey: "whitelist"}

type WhiteList struct {
	sync.Mutex
	DBKey string `bson:"_id"`
	List  []string
}

func Http_whitelist_add(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v := q.Get("val")

	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	G_WhiteList.Add(v)
	w.Write(common.S2B("ok"))
	gamelog.Info("Http_whitelist_add: %v", r.Form)
}
func Http_whitelist_del(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v := q.Get("val")

	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	G_WhiteList.Del(v)
	w.Write(common.S2B("ok"))
	gamelog.Info("Http_whitelist_del: %v", r.Form)
}

// ------------------------------------------------------------
func (self *WhiteList) Have(v string) bool {
	self.Lock()
	defer self.Unlock()
	for _, vv := range self.List {
		if vv == v {
			return true
		}
	}
	return false
}
func (self *WhiteList) Add(v string) {
	self.Lock()
	defer self.Unlock()
	for _, vv := range self.List {
		if vv == v {
			return
		}
	}
	self.List = append(self.List, v)
	self.UpdateDB()
}
func (self *WhiteList) Del(v string) {
	self.Lock()
	defer self.Unlock()
	for i, vv := range self.List {
		if vv == v {
			self.List = append(self.List[:i], self.List[i+1:]...)
			self.UpdateDB()
			return
		}
	}
}

// ------------------------------------------------------------
func (self *WhiteList) UpdateDB() { dbmgo.UpdateId(dbmgo.KTableArgs, self.DBKey, self) }
func (self *WhiteList) InitDB() {
	if ok, _ := dbmgo.Find(dbmgo.KTableArgs, "_id", self.DBKey, self); !ok {
		dbmgo.Insert(dbmgo.KTableArgs, self)
	}
}
