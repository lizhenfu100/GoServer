package account

import (
	"common"
	"common/format"
	"common/std/sign"
	"common/tool/email"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"netConfig/meta"
	"strconv"
	"time"
)

// -------------------------------------
// 绑定信息到账号
func Rpc_center_bind_info(req, ack *common.NetPack) {
	name := req.ReadString()
	passwd := req.ReadString()
	key := req.ReadString()
	val := req.ReadString()
	force := req.ReadBool()

	errcode := BindInfoToAccount(name, passwd, key, val, force)
	ack.WriteUInt16(errcode)
}
func Rpc_center_get_account_by_bind_info(req, ack *common.NetPack) { //拿到账号名，client本地保存
	val := req.ReadString()
	key := req.ReadString()
	passwd := req.ReadString()

	account := GetAccountByBindInfo(key, val)
	if account != nil {
		ack.WriteUInt16(err.Account_none)
	} else if !account.CheckPasswd(passwd) {
		ack.WriteUInt16(err.Passwd_err)
	} else {
		ack.WriteUInt16(err.Success)
		ack.WriteString(account.Name)
	}
}
func Rpc_center_get_bind_info(req, ack *common.NetPack) {
	name := req.ReadString()
	key := req.ReadString()

	if account := GetAccountByName(name); account != nil {
		if v, ok := account.BindInfo[key]; ok {
			ack.WriteString(v)
			return
		}
	}
	ack.WriteString("")
}

// -------------------------------------
// 辅助函数
func BindInfoToAccount(name, passwd, k, v string, force bool) (errcode uint16) {
	if account := GetAccountByName(name); account == nil {
		errcode = err.Account_none
	} else if !account.CheckPasswd(passwd) {
		errcode = err.Passwd_err
	} else if account.IsForbidden {
		errcode = err.Account_forbidden
	} else if !format.CheckBindValue(k, v) {
		errcode = err.BindInfo_format_err
	} else if GetAccountByBindInfo(k, v) != nil {
		errcode = err.BindInfo_had_been_bound //this_value_already_bind_to_account
	} else {
		if force { //强制绑定，须验证
			errcode = err.Success
			account.forceBind(k, v)
		} else if _, ok := account.BindInfo[k]; ok {
			errcode = err.Account_had_been_bound
		} else {
			errcode = err.Success
			account.bind(k, v)
		}
	}
	return
}
func GetAccountByBindInfo(k, v string) *TAccount {
	//FIXME: 数据多了，这样没加索引的找太慢了
	//客户端通过绑定信息查到账号后，将账号名保存至本地，之后用账户名登录
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	account := new(TAccount)
	if ok, _ := dbmgo.Find(KDBTable, dbkey, v, account); ok {
		return account
	}
	return nil
}

func (self *TAccount) bind(k, v string) {
	self.BindInfo[k] = v
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{dbkey: v}})
}
func (self *TAccount) forceBind(k, v string) {
	switch k {
	case "email":
		//1、创建url
		httpAddr := fmt.Sprintf("http://%s:%d/bind_info_force",
			meta.G_Local.OutIP, meta.G_Local.Port())
		u, _ := url.Parse(httpAddr)
		q := u.Query()
		//2、写入参数
		q.Set("name", self.Name)
		q.Set("k", k)
		q.Set("v", v)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(self.Name+k+v+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		email.SendMail("Verify Email", v, u.String(), "")
	default:
		self.bind(k, v)
	}
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	k := q.Get("k")
	v := q.Get("v")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()

	if sign.CalcSign(name+k+v+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if account := GetAccountByName(name); account == nil {
		ack = "Error: Account_none"
	} else {
		account.bind(k, v)
		ack = "Bind info ok"
	}
}
