package web

import (
	"common/format"
	"common/timer"
	"common/tool/sms"
	"dbmgo"
	"encoding/json"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

const (
	KDBUser  = "user"
	Need_SMS = false
)

// db.user.insert({_id:"15102165079", name:"zhoumf", pwd:"Zmf_890104"})
// db.log.find().sort({_id:-1})
type User struct {
	Phone     string `bson:"_id" json:"username"`
	Name      string
	Pwd       string `json:"password"`
	Code      string `bson:"-"`
	Comfirpwd string `bson:"-" json:"comfirpwd"`
}

//解决ajax的跨域访问问题
func InitAjax(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "x-requested-with,content-type")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	r.Header.Set("Content-Type", "application/json;charset=UTF-8")
}
func SetCookies(w http.ResponseWriter, key, val string) {
	cookie := http.Cookie{Name: key, Value: val, Path: "/"}
	http.SetCookie(w, &cookie)
}
func CheckCookies(r *http.Request) bool {
	if p, err := r.Cookie("username"); err != nil || p.Value == "" {
		return false
	} else {
		dbmgo.Log(r.URL.Path, p.Value, r.URL.RawQuery)
		return true
	}
}
func JsonRespone(respone interface{}, w http.ResponseWriter) {
	if buf, e := json.Marshal(respone); e == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	}
}
func Http_login(w http.ResponseWriter, r *http.Request) {
	InitAjax(&w, r)
	var user User
	userIp := strings.Split(r.RemoteAddr, ":")[0]
	buf, _ := ioutil.ReadAll(r.Body)
	if json.Unmarshal(buf, &user) != nil {
		JsonRespone("未知错误", w)
		onLoginFail(userIp, user.Phone)
	} else if inBlackList(userIp, user.Phone) {
		JsonRespone("改账号已被封禁，请连续管理员", w)
	} else if pwd := user.Pwd; Need_SMS && !sms.CheckCode(user.Phone, user.Code) {
		JsonRespone("验证码错误，请重新输入", w)
		onLoginFail(userIp, user.Phone)
	} else if ok, _ := dbmgo.Find(KDBUser, "_id", user.Phone, &user); !ok {
		JsonRespone("不存在该用户", w)
		onLoginFail(userIp, user.Phone)
	} else if user.Pwd != pwd {
		JsonRespone("密码错误", w)
		onLoginFail(userIp, user.Phone)
	} else {
		SetCookies(w, "username", user.Name+" "+userIp)
		JsonRespone("成功", w)
	}
}
func Http_reset_pwd(w http.ResponseWriter, r *http.Request) {
	InitAjax(&w, r)
	var user User
	buf, _ := ioutil.ReadAll(r.Body)
	if json.Unmarshal(buf, &user) != nil {
		JsonRespone("未知错误", w)
	} else if comfirpwd, pwd := user.Comfirpwd, user.Pwd; !format.CheckPasswdEx(comfirpwd) {
		JsonRespone("密码至少8位，须包含：大小写字母、数字、特殊字符", w)
	} else if ok, err := dbmgo.Find(KDBUser, "_id", user.Phone, &user); err == nil && ok == false {
		JsonRespone("不存在该用户", w)
	} else if pwd != user.Pwd {
		JsonRespone("密码错误", w)
	} else if !dbmgo.UpdateIdSync(KDBUser, user.Phone, bson.M{"$set": bson.M{"pwd": comfirpwd}}) {
		JsonRespone("密码重置失败", w)
	} else {
		JsonRespone("成功", w)
	}
}
func Http_send_sms(w http.ResponseWriter, r *http.Request) {
	InitAjax(&w, r)
	var phone string
	buf, _ := ioutil.ReadAll(r.Body)
	if json.Unmarshal(buf, &phone) != nil {
		JsonRespone("请输入有效的手机号", w)
	} else if sms.SendCode(phone) != 1 {
		JsonRespone("发送失败", w)
	} else {
		JsonRespone("发送成功", w)
	}
}

// ------------------------------------------------------------
// 连续3次登录失败，封ip、账号
var (
	_blackList sync.Map
	_loginFreq = timer.NewFreq(3, 10)
)

func inBlackList(ip, user string) bool {
	if _, ok := _blackList.Load(ip); ok {
		return true
	} else if _, ok = _blackList.Load(user); ok {
		return true
	}
	return false
}
func onLoginFail(ip, user string) {
	if !_loginFreq.Check(ip) {
		_blackList.Store(ip, true)
		gamelog.Error("Forbid: %s", ip)
	} else if !_loginFreq.Check(user) {
		_blackList.Store(user, true)
		gamelog.Error("Forbid: %s", user)
	}
}
