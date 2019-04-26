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
	k := req.ReadString()
	v := req.ReadString()
	force := req.ReadBool()

	errcode := err.Unknow_error
	if account := GetAccountByName(name); account == nil {
		errcode = err.Account_none
	} else if !account.CheckPasswd(passwd) {
		errcode = err.Passwd_err
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
	ack.WriteUInt16(errcode)
}
func Rpc_center_get_account_by_bind_info(req, ack *common.NetPack) { //拿到账号名，client本地保存
	val := req.ReadString()
	key := req.ReadString()
	passwd := req.ReadString()

	if account := GetAccountByBindInfo(key, val); account == nil {
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
func (self *TAccount) bind(k, v string) {
	self.BindInfo[k] = v
	dbkey := fmt.Sprintf("bindinfo.%s", k)
	dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{dbkey: v}})
}
func (self *TAccount) forceBind(k, v string) {
	switch k {
	case "email":
		//1、创建url
		u, _ := url.Parse(fmt.Sprintf("http://%s:%d/bind_info_force",
			meta.G_Local.OutIP, meta.G_Local.Port()))
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
		email.SendMail2("Verify Email", v, u.String(), "")
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
