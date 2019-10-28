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
func Rpc_center_bind_info(req, ack *common.NetPack) {
	str := req.ReadString()
	passwd := req.ReadString()
	k := req.ReadString()
	v := req.ReadString()
	force := req.ReadBool()
	sign.Decode(&passwd)

	errcode, ptr := GetAccount(str, passwd)
	if errcode == err.Success {
		if !format.CheckBindValue(k, v) {
			errcode = err.BindInfo_format_err
		} else if GetAccountByBindInfo(k, v) != nil {
			errcode = err.BindInfo_already_in_use
		} else if k == "email" && ptr.IsValidEmail > 0 {
			errcode = err.Is_forbidden
		} else {
			if _, ok := ptr.BindInfo[k]; !ok {
				errcode = ptr.bind(k, v)
			} else if force { //强制绑定，须验证
				errcode = ptr.forceBind(k, v)
			} else {
				errcode = err.Account_had_been_bound
			}
		}
	}
	ack.WriteUInt16(errcode)
}
func Rpc_center_isvalid_bind_info(req, ack *common.NetPack) {
	val := req.ReadString()
	typ := req.ReadString()
	if p := GetAccountByBindInfo(typ, val); p == nil {
		ack.WriteUInt16(err.Not_found)
	} else if typ == "email" && p.IsValidEmail <= 0 {
		ack.WriteUInt16(err.Invalid)
	} else if typ == "phone" && p.IsValidPhone <= 0 {
		ack.WriteUInt16(err.Invalid)
	} else {
		ack.WriteUInt16(err.Success)
	}
}

// ------------------------------------------------------------
// 辅助函数
func (self *TAccount) bind(k, newVal string) uint16 {
	if GetAccountByBindInfo(k, newVal) != nil {
		return err.BindInfo_already_in_use
	}
	DelCache(self) //更新缓存
	dbmgo.Log("Change_"+k, self.BindInfo[k], newVal)
	self.BindInfo[k] = newVal
	AddCache(self)
	dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"bindinfo." + k: newVal}})
	return err.Success
}
func (self *TAccount) forceBind(k, v string) uint16 {
	switch k {
	case "email":
		//1、创建url
		u, _ := url.Parse(fmt.Sprintf("http://%s:%d/bind_info_force",
			meta.G_Local.OutIP, meta.G_Local.Port()))
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
	aid := q.Get("aid")
	k := q.Get("k")
	v := q.Get("v")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)
	aids, _ := strconv.Atoi(aid)
	accountId := uint32(aids)

	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()

	if sign.CalcSign(aid+k+v+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", "")
	} else if time.Now().Unix()-timeFlag > 3600 {
		ack, _ = email.Translate("Error: url expire", "")
	} else if account := GetAccountById(accountId); account == nil {
		ack, _ = email.Translate("Error: Account_none", "")
	} else {
		if account.bind(k, v) == err.Success {
			if k == "email" {
				account.verifyEmailOK()
			}
			ack, _ = email.Translate("Bind info ok", "")
		} else {
			ack = "Error: BindInfo_already_in_use"
		}
	}
}

// ------------------------------------------------------------
// 邮箱验证
func (self *TAccount) verifyEmailOK() {
	self.IsValidEmail = 1
	dbmgo.UpdateId(KDBTable, self.AccountID, bson.M{"$set": bson.M{"isvalidemail": 1}})
}
func Http_verify_email(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("email")
	flag := q.Get("flag")
	language := q.Get("language")
	timeFlag, _ := strconv.ParseInt(flag, 10, 64)

	ack := "Error: unknown"
	defer func() {
		w.Write(common.S2B(ack))
	}()
	if sign.CalcSign(addr+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", language)
	} else if time.Now().Unix()-timeFlag > 7*24*3600 {
		ack, _ = email.Translate("Error: url expire", language)
	} else if p := GetAccountByBindInfo("email", addr); p == nil {
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
	defer func() {
		w.Write(common.S2B(ack))
	}()
	if sign.CalcSign(addr+passwd+flag) != q.Get("sign") {
		ack, _ = email.Translate("Error: sign failed", language)
	} else if time.Now().Unix()-timeFlag > 7*24*3600 {
		ack, _ = email.Translate("Error: url expire", language)
	} else if sign.Decode(&passwd); !format.CheckPasswd(passwd) {
		ack, _ = email.Translate("Error: Passwd_format_err", language)
	} else if _, p := NewAccountInDB(passwd, "email", addr); p == nil {
		ack, _ = email.Translate("Error: Account_reg_fail", language)
	} else {
		p.verifyEmailOK()
		ack, _ = email.Translate("Account reg ok", language)
	}
}
