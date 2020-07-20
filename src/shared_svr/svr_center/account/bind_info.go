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

// ------------------------------------------------------------
// 绑定信息到账号
func Rpc_center_bind_info(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	pwd := req.ReadString()
	k := req.ReadString()
	v := req.ReadString()
	force := req.ReadBool()
	sign.Decode(&str, &pwd)
	errcode, ptr := GetAccount(str, pwd)
	if ptr != nil {
		if !format.CheckBindValue(k, v) {
			errcode = err.BindInfo_format_err
		} else if k == "email" && ptr.IsValidEmail == 1 {
			errcode = err.Is_forbidden
		} else if e, _ := GetAccountByBindInfo(k, v); e == err.Success {
			errcode = err.BindInfo_already_in_use
		} else if e == err.Not_found {
			if _, ok := ptr.BindInfo[k]; !ok {
				errcode = ptr.bind(k, v)
			} else if force { //强制绑定，须验证
				errcode = ptr.bindVerify(k, v)
			} else {
				errcode = err.Account_had_been_bound
			}
		} else {
			errcode = e
		}
	}
	ack.WriteUInt16(errcode)
}
func Rpc_center_isvalid_bind_info(req, ack *common.NetPack, _ common.Conn) {
	v := req.ReadString()
	k := req.ReadString()
	if e, p := GetAccountByBindInfo(k, v); p == nil {
		ack.WriteUInt16(e)
	} else if k == "email" && p.IsValidEmail <= 0 {
		ack.WriteUInt16(err.Invalid)
	} else if k == "phone" && p.BindInfo[k] == "" {
		ack.WriteUInt16(err.Invalid)
	} else {
		ack.WriteUInt16(err.Success)
	}
}

// ------------------------------------------------------------
// 辅助函数
func (self *TAccount) bind(typ, newVal string) uint16 {
	if e, p := GetAccountByBindInfo(typ, newVal); p != nil {
		return err.BindInfo_already_in_use
	} else if e == err.Not_found {
		CacheDel(self)
		dbmgo.Log("Change_"+typ, self.BindInfo[typ], newVal)
		self.BindInfo[typ] = newVal
		CacheAdd(self)
		dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"bindinfo." + typ: newVal}})
		return err.Success
	} else {
		return e
	}
}
func (self *TAccount) bindVerify(k, v string) uint16 {
	switch k {
	case "email":
		//1、创建url
		u, _ := url.Parse(fmt.Sprintf("http://%s:%d/bind_info_force",
			meta.G_Local.OutIP, meta.G_Local.HttpPort))
		q := u.Query()
		//2、写入参数
		aid := strconv.Itoa(int(self.AccountID))
		q.Set("aid", aid)
		q.Set("k", k)
		q.Set("v", v)
		flag := strconv.FormatInt(time.Now().Unix(), 10)
		q.Set("flag", flag)
		q.Set("sign", sign.CalcSign(aid+k+v+flag))
		//3、生成完整url
		u.RawQuery = q.Encode()
		email.SendMail("Verify Email", v, u.String(), "")
		return err.Success
	default:
		return self.bind(k, v)
	}
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	k, v := q.Get("k"), q.Get("v")
	aid := q.Get("aid")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)
	aids, _ := strconv.Atoi(aid)
	accountId := uint32(aids)

	ack := "Error: unknown"
	defer func() {
		if k == "email" {
			ack, _ = email.Translate(ack, "")
		}
		w.Write(common.S2B(ack))
	}()
	if sign.CalcSign(aid+k+v+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack = "Error: url expire"
	} else if _, p := GetAccountById(accountId); p == nil {
		ack = "Error: Account_none"
	} else if p.bind(k, v) == err.Success {
		if k == "email" {
			p.verifyEmailOK()
		}
		ack = "Bind info ok"
	} else {
		ack = "Error: BindInfo_already_in_use"
	}
}

// ------------------------------------------------------------
// 邮箱验证 【不同游戏间验证，不保障强一致性，遍历各个游戏太挫了】
func (self *TAccount) verifyEmailOK() {
	if self.IsValidEmail == 0 {
		self.IsValidEmail = 1
		dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"isvalidemail": 1}})
		CacheDel(self)
	}
}
func Http_verify_email(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("email")
	flag := q.Get("flag")
	language := q.Get("language")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	ack := "Error: unknown"
	defer func() { w.Write(common.S2B(ack)) }()
	if sign.CalcSign(addr+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", language)
	} else if time.Now().Unix()-timeFlag > 7*24*3600 {
		ack, _ = email.Translate("Error: url expire", language)
	} else if _, p := GetAccountByBindInfo("email", addr); p == nil {
		ack, _ = email.Translate("Error: Account_none", language)
	} else {
		p.verifyEmailOK()
		ack, _ = email.Translate("Verify OK", language)
	}
}
func Http_reg_account_by_email(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("email")
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	language := q.Get("language")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	ack := "Error: unknown"
	defer func() { w.Write(common.S2B(ack)) }()
	if sign.CalcSign(addr+passwd+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", language)
	} else if time.Now().Unix()-timeFlag > 7*24*3600 {
		ack, _ = email.Translate("Error: url expire", language)
	} else if !format.CheckPasswd(passwd) {
		ack, _ = email.Translate("Error: Passwd_format_err", language)
	} else if _, p := NewAccountInDB(passwd, "email", addr); p == nil {
		ack, _ = email.Translate("Error: Account_reg_fail", language)
	} else {
		p.verifyEmailOK()
		ack, _ = email.Translate("Account reg ok", language)
	}
}
