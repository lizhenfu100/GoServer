package main

import (
	"bytes"
	"common"
	"common/copy"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/http"
	"netConfig/meta"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	kDBTable     = "afk"
	kPassword    = "chillyroom_afk_*"
	kFileDirRoot = "html/AFK/"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
}

type Afk_req struct {
	ID         uint32 `bson:"_id"`
	Name       string
	Type       string
	Reason     string
	AfkTime    string //请假时间
	RepairTime string //补班时间
	Time       int64  //时间戳
}

func (self *Afk_req) check() (err string) {
	self.Reason = strings.Replace(self.Reason, " ", "", -1)
	self.AfkTime = strings.Replace(self.AfkTime, " ", "", -1)
	self.RepairTime = strings.Replace(self.RepairTime, " ", "", -1)

	if self.Name == "" {
		err = "失败：姓名缺失"
	}
	if self.AfkTime == "" {
		err = "失败：未注明请假时间"
	}
	if self.Type == "日调" && self.RepairTime == "" {
		err = "日调失败：未注明补班时间"
	}
	if self.Type == "补班" && self.RepairTime == "" {
		err = "补班失败：未注明补班时间"
	}
	if self.Type == "调班" && self.RepairTime == "" {
		err = "调班失败：未注明补班时间"
	}

	if ok, _ := dbmgo.FindEx(kDBTable, bson.M{"name": self.Name, "afktime": self.AfkTime}, &Afk_req{}); ok {
		err = "失败：重复请假：" + self.AfkTime
	}
	return
}
func (self *Afk_req) format() string {
	return fmt.Sprintf("#%s %s %s %s\n\n\n%s 申请成功，请复制上述结果到工作群",
		self.Type, self.Reason, self.AfkTime, self.RepairTime, self.Name)
}

func replyHtml(w http.ResponseWriter, name string) {
	if t, err := template.ParseFiles(kFileDirRoot + name); err != nil {
		fmt.Fprintf(w, "parse template error: %s", err.Error())
	} else {
		t.Execute(w, meta.G_Local)
	}
}

func Http_afk(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		replyHtml(w, "afk.html")
	} else {
		r.ParseForm()

		//反射解析
		var req Afk_req
		copy.CopyForm(&req, r.Form)

		//! 创建回复
		ack := "失败"
		defer func() {
			w.Write(common.S2B(ack))
		}()

		if r.Form.Get("passwd") != kPassword {
			ack = "密码错误"
			return
		}

		if ack = req.check(); ack == "" {
			req.ID = dbmgo.GetNextIncId("AfkId")
			req.Time = time.Now().Unix()
			dbmgo.Insert(kDBTable, &req)
			ack = req.format()
		}
	}
}
func Http_del(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		replyHtml(w, "del.html")
	} else {
		r.ParseForm()

		//! 创建回复
		ack := "失败，无此信息"
		defer func() {
			w.Write(common.S2B(ack))
		}()

		if r.Form.Get("passwd") != kPassword {
			ack = "密码错误"
			return
		}

		if dbmgo.RemoveAllSync(kDBTable, bson.M{
			"name":    r.Form.Get("name"),
			"afktime": r.Form.Get("afktime"),
		}) {
			ack = "成功删除"
		}
	}
}
func Http_count_one(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		replyHtml(w, "count.html")
	} else {
		r.ParseForm()

		//! 创建回复
		ack := "失败，无此信息"
		defer func() {
			w.Write(common.S2B(ack))
		}()

		if r.Form.Get("passwd") != kPassword {
			ack = "密码错误"
			return
		}

		name := r.Form.Get("name")
		begin := getTime(r.Form.Get("begin"), "-")
		end := getTime(r.Form.Get("end"), "-")
		if name == "" {
			ack = "失败：姓名缺失"
			return
		}
		if begin == nil || end == nil {
			ack = "失败，日期错误，正确格式2018-1-15"
			return
		}
		ack = queryDuringTime(name, begin, end)
	}
}
func Http_count_all(req, ack *common.NetPack) {
	begin := getTime(req.ReadString(), ".")
	end := getTime(req.ReadString(), ".")
	if begin == nil || end == nil {
		ack.WriteString("失败，日期错误，正确格式2018.1.15")
	} else {
		ack.WriteString("ok")
		ack.WriteBuf(common.S2B(queryDuringTime("", begin, end)))
	}
}

// ------------------------------------------------------------
func getTime(date, sep string) *time.Time {
	v := strings.Split(date, sep)
	if len(v) < 3 {
		return nil
	}
	year, _ := strconv.Atoi(v[0])
	month, _ := strconv.Atoi(v[1])
	day, _ := strconv.Atoi(v[2])
	if year == 0 || month == 0 || day == 0 {
		return nil
	}
	ret := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())
	return &ret
}
func queryDuringTime(name string, begin, end *time.Time) string {
	var list infoSlice
	if name == "" {
		//删除：过期一年的
		dbmgo.RemoveAllSync(kDBTable, bson.M{
			"type": bson.M{"$ne": "年假"},
			"time": bson.M{"$lt": time.Now().Unix() - 365*24*3600},
		})
		dbmgo.FindAll(kDBTable, bson.M{
			"time": bson.M{"$gte": begin.Unix(), "$lt": end.Unix()},
		}, &list)
	} else {
		dbmgo.FindAll(kDBTable, bson.M{
			"name": name,
			"time": bson.M{"$gte": begin.Unix(), "$lt": end.Unix()},
		}, &list)
	}

	//打包回复
	sort.Sort(list)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("【%d.%d-%d.%d统计结果】\n",
		begin.Month(), begin.Day(),
		end.Month(), end.Day()))
	curName, curTypeLv := "", 0
	for _, v := range list {
		if curName != v.Name {
			curName = v.Name
			curTypeLv = 0
			buf.WriteString("\n" + curName + "：\n")
		}
		if curTypeLv != typeLv(v.Type) {
			if curTypeLv != 0 {
				buf.WriteString("\n")
			}
			curTypeLv = typeLv(v.Type)
		}
		buf.WriteString("    #")
		buf.WriteString(v.Type)
		buf.WriteString(" ")
		buf.WriteString(v.Reason)
		buf.WriteString(" ")
		buf.WriteString(v.AfkTime)
		buf.WriteString(" ")
		buf.WriteString(v.RepairTime)
		buf.WriteString("\n")
	}
	return buf.String()
}

// ------------------------------------------------------------
type infoSlice []Afk_req

func (c infoSlice) Len() int      { return len(c) }
func (c infoSlice) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c infoSlice) Less(i, j int) bool {
	if c[i].Name == c[j].Name {
		return typeLv(c[i].Type) > typeLv(c[j].Type)
	} else {
		return c[i].Name < c[j].Name
	}
}
func typeLv(typ string) int {
	if v, ok := g_TypeLv[typ]; ok {
		return v
	}
	return 0
}
