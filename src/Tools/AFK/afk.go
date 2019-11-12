package afk

import (
	"Tools/AFK/qqmsg"
	"bytes"
	"common"
	"common/copy"
	"common/file"
	"common/std"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"netConfig/meta"
	"nets"
	"sort"
	"strconv"
	"text/template"
	"time"
)

const (
	kDBTable     = "afk"
	kFileDirRoot = "html/AFK/"
	kTimeLayout  = "2006-01-02"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir(kFileDirRoot)))
}

type Afk_req struct {
	ID     uint32 `bson:"_id"`
	Name   string
	Type   string
	Reason string
	Date   string //2006-01-02
	Hour   int

	//辅助统计
	Year  int //哪一年
	Month int //几月
	Week  int //本年的第几周
}

func (self *Afk_req) init() {
	t := time.Now()
	if self.Type != "周补" && self.Type != "事补" {
		t = s2t(self.Date)
	}
	self.Year, self.Week = t.ISOWeek()
	self.Month = int(t.Month())
}
func (self *Afk_req) print() string {
	return fmt.Sprintf("%s：#%s %s %s %dh\n", self.Name, self.Type, self.Reason, self.Date, self.Hour)
}
func Http_afk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var req Afk_req
	copy.CopyForm(&req, r.Form)
	req.init()

	ack := "失败，请稍后重试"
	defer func() {
		w.Write(common.S2B(ack))
	}()
	if err := req.check(); err == "" {
		req.ID = dbmgo.GetNextIncId("AfkId")
		if dbmgo.InsertSync(kDBTable, &req) {
			ack = req.print()
			qqmsg.G.Add(ack)
		}
	} else {
		ack = err
	}
}
func Http_del(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.Form.Get("id"))
	ack := fmt.Sprintf("失败，无此信息 id(%d)", id)
	defer func() {
		w.Write(common.S2B(ack))
	}()
	var v Afk_req
	if ok, _ := dbmgo.Find(kDBTable, "_id", id, &v); ok {
		if dbmgo.RemoveOneSync(kDBTable, bson.M{"_id": id}) {
			ack = "删除：\n    "
			ack += v.print()
			qqmsg.G.Add(ack)
		}
	}
}
func Http_my_afk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	begin := r.Form.Get("begin")
	end := r.Form.Get("end")

	var list infoSlice
	dbmgo.FindAll(kDBTable, bson.M{"name": name, "date": bson.M{"$gte": begin, "$lte": end}}, &list)
	t, _ := template.New("").Parse(`<html>
<body bgcolor="white">
    <table border="1" cellpadding="5" cellspacing="0">
        <tr>
            <th>类型</th><th>原因</th><th>Date</th><th>Hour</th>
        </tr>
{{range $_, $v := .}}
        <tr>
            <td>{{$v.Type}}</td><td>{{$v.Reason}}</td><td>{{$v.Date}}</td><td>{{$v.Hour}}</td>
			<td>
				<form action="http://www.chillyroom.com/api/del" method="post">
                	<input type="hidden" name="id" value={{$v.ID}}><br>
                	<input type="submit" value="删除">
            	</form>
			</td>
        </tr>
{{end}}
    </table>
</body>
</html>`)
	t.Execute(w, list)
}
func Http_count(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	begin := r.Form.Get("begin")
	end := r.Form.Get("end")

	t, _ := template.New("").Parse(`<html>
<body bgcolor="white">
{{.}}
</body>
</html>`)
	t.Execute(w, query(name, begin, end))
}

// ------------------------------------------------------------
func (self *Afk_req) check() (err string) {
	if g_names.Index(self.Name) < 0 {
		err = "失败：非名单人员"
	}
	if self.Date == "" || self.Hour == 0 {
		err = "失败：未注明请假时间"
	}
	switch self.Type {
	case "周调":
		if calcWeekHours(self.Name, self.Year, self.Week)+self.Hour > 4 {
			err = "失败：周调，每周最多4小时"
		}
	case "周补":
		wday := s2t(self.Date).Weekday()
		if wday != time.Saturday {
			err = "失败：周补，仅可在周六"
		}
	case "事调":
		if calcMonthCnt(self.Name, self.Year, self.Month) >= 2 {
			err = "失败：事调，每月最多2次"
		}
	}
	return
}
func s2t(s string) time.Time {
	t, _ := time.ParseInLocation(kTimeLayout, s, time.Local)
	return t
}
func calcWeekHours(name string, year, week int) int { //计算周调总小时数
	list := infoSlice{}
	dbmgo.FindAll(kDBTable, bson.M{"name": name, "type": "周调", "year": year, "week": week}, &list)
	sum := 0
	for _, v := range list {
		sum += v.Hour
	}
	return sum
}
func calcMonthCnt(name string, year, month int) int { //计算事调次数
	list := infoSlice{}
	dbmgo.FindAll(kDBTable, bson.M{"name": name, "type": "事调", "year": year, "month": month}, &list)
	return len(list)
}

// ------------------------------------------------------------
func query(name string, begin, end string) string {
	var list infoSlice
	if name == "" {
		dbmgo.RemoveAllSync(kDBTable, bson.M{"year": bson.M{"$lt": time.Now().Year() - 1}})
		dbmgo.FindAll(kDBTable, bson.M{"date": bson.M{"$gte": begin, "$lte": end}}, &list)
	} else {
		dbmgo.FindAll(kDBTable, bson.M{"name": name, "date": bson.M{"$gte": begin, "$lte": end}}, &list)
	}
	sort.Sort(list)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("【%s ~ %s统计结果】<br>", begin, end))
	curName, curTypeLv, person := "", 0, infoSlice{}
	for _, v := range list {
		if curName != v.Name {
			curName = v.Name
			curTypeLv = 0
			countPerson(person, &buf)
			person = person[:0] //换人后清空
			buf.WriteString("<br>" + curName + "：<br>")
		}
		person = append(person, v)
		if curTypeLv != typeLv1(v.Type) {
			if curTypeLv != 0 {
				buf.WriteString("<br>")
			}
			curTypeLv = typeLv1(v.Type)
		}
		buf.WriteString("&nbsp;&nbsp;&nbsp;&nbsp;")
		if v.Type == "周补" || v.Type == "事补" {
			buf.WriteString("&nbsp;&nbsp;&nbsp;&nbsp;")
		}
		if v.Type == "事假" || v.Type == "特殊" {
			buf.WriteString(fmt.Sprintf("<font color=#FF0000>#%s %s %s %dh</font><br>",
				v.Type, v.Reason, v.Date, v.Hour))
		} else {
			buf.WriteString(fmt.Sprintf("#%s %s %s %dh<br>", v.Type, v.Reason, v.Date, v.Hour))
		}
	}
	if len(person) > 0 {
		countPerson(person, &buf)
	}
	buf.WriteString(countNoneAfk(list))
	return buf.String()
}
func countPerson(person infoSlice, buf *bytes.Buffer) {
	weekHour1, weekHour2, monthHour1, monthHour2 := 0, 0, 0, 0
	yearHour := 0
	for _, v := range person {
		switch v.Type {
		case "周调":
			weekHour1 += v.Hour
		case "周补":
			weekHour2 += v.Hour
		case "事调":
			monthHour1 += v.Hour
		case "事补":
			monthHour2 += v.Hour
		case "年假":
			yearHour += v.Hour
		}
	}
	if weekHour1-weekHour2 != 0 || monthHour1-monthHour2 != 0 {
		buf.WriteString(fmt.Sprintf("<br>&nbsp;&nbsp;&nbsp;&nbsp;<font color=#FF0000>周调未补(%dh)，事调未补(%dh)</font><br>",
			weekHour1-weekHour2, monthHour1-monthHour2))
	}
	if yearHour != 0 {
		buf.WriteString(fmt.Sprintf("<br>&nbsp;&nbsp;&nbsp;&nbsp;<font color=#FF0000>年假共计(%dd%dh)</font><br>",
			yearHour/8, yearHour%8))
	}
}
func countNoneAfk(list infoSlice) string {
	ifafk := func(name string) bool {
		for _, v := range list {
			if v.Name == name {
				return true
			}
		}
		return false
	}
	var ret std.Strings
	for _, v := range g_names {
		if !ifafk(v) {
			ret.Add(v)
		}
	}
	var buf bytes.Buffer
	buf.WriteString("<font color=#FF0000><br>全勤名单：</font>")
	for i, v := range ret {
		if i%6 == 0 {
			buf.WriteString("<br>&nbsp;&nbsp;&nbsp;&nbsp;")
		}
		buf.WriteString(v)
		buf.WriteString("&nbsp;")
	}
	return buf.String()
}

// ------------------------------------------------------------
type infoSlice []Afk_req

func (c infoSlice) Len() int      { return len(c) }
func (c infoSlice) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c infoSlice) Less(i, j int) bool {
	if c[i].Name == c[j].Name {
		if typeLv1(c[i].Type) == typeLv1(c[j].Type) {
			return typeLv2(c[i].Type) < typeLv2(c[j].Type)
		} else {
			return typeLv1(c[i].Type) < typeLv1(c[j].Type)
		}
	} else {
		return c[i].Name < c[j].Name
	}
}
func typeLv1(typ string) int {
	if v, ok := g_TypeLv[typ]; ok {
		return v[0]
	}
	return 0
}
func typeLv2(typ string) int {
	if v, ok := g_TypeLv[typ]; ok {
		return v[1]
	}
	return 0
}

// ------------------------------------------------------------
func Init() {
	nets.RegHttpHandler(map[string]nets.HttpHandle{
		"/afk":      Http_afk,
		"/del":      Http_del,
		"/count":    Http_count,
		"/my_afk":   Http_my_afk,
		"/afk_msgs": qqmsg.Http_msgs,
	})
	file.TemplateDir(meta.G_Local, kFileDirRoot+"template/", kFileDirRoot, ".html")
}
