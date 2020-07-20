package account

import (
	"common"
	"common/format"
	"common/std/sign"
	"dbmgo"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/account/gameInfo"
	"sync/atomic"
	"time"
)

func Rpc_check_identity(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString() //email、name、phone
	pwd := req.ReadString()
	gameName := req.ReadString()
	if e, p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(e)
	} else if !p.CheckPasswd(pwd) {
		ack.WriteUInt16(err.Account_mismatch_passwd)
	} else {
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&p.LoginTime, timeNow)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		ack.WriteUInt16(err.Success)
		(&gameInfo.TAccountClient{ //1、先回复，Client账号信息
			p.AccountID,
			p.IsValidEmail,
			p.BindInfo,
		}).DataToBuf(ack)
		if v, ok := p.GameInfo[gameName]; ok { //2、附带的游戏数据，可能有的游戏空的
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_account_reg2(req, ack *common.NetPack, _ common.Conn) {
	Rpc_center_account_reg(req, ack, nil)
}
func Rpc_center_platform_reg(req, ack *common.NetPack, _ common.Conn) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	if uid == "" || pf_id == "" {
		ack.WriteUInt16(err.BindInfo_format_err)
	} else {
		sign.Decode(&uid)
		e, _ := NewAccountInDB("", "name", pf_id+"_"+uid)
		ack.WriteUInt16(e)
	}
}
func Rpc_center_reg_if2(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	sign.Decode(&str)
	e, _ := GetAccountByBindInfo(typ, str)
	ack.WriteUInt16(e)
}
func Rpc_center_set_game_route2(req, ack *common.NetPack, _ common.Conn) {
	Rpc_center_set_game_route(req, ack, nil)
}
func Rpc_center_set_game_json2(req, ack *common.NetPack, _ common.Conn) {
	Rpc_center_set_game_json(req, ack, nil)
}
func Rpc_center_game_info2(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	gameName := req.ReadString()
	sign.Decode(&str)
	if e, p := GetAccountByBindInfo(typ, str); e == err.Unknow_error {
		ack.WriteUInt16(e)
	} else if ack.WriteUInt16(err.Success); p != nil {
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteString(v.JsonData)
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_player_addr2(req, ack *common.NetPack, _ common.Conn) {
	Rpc_center_player_login_addr(req, ack, nil)
}
func Rpc_center_bind_info2(req, ack *common.NetPack, _ common.Conn) {
	str := req.ReadString()
	typ := req.ReadString()
	pwd := req.ReadString()
	k := req.ReadString()
	v := req.ReadString()
	sign.Decode(&str, &pwd)
	errcode, ptr := GetAccountByBindInfo(typ, str)
	if ptr != nil {
		if errcode = err.Unknow_error; !ptr.CheckPasswd(pwd) {
			errcode = err.Account_mismatch_passwd
		} else if !format.CheckBindValue(k, v) {
			errcode = err.BindInfo_format_err
		} else if k == "email" && ptr.IsValidEmail == 1 {
			errcode = err.Is_forbidden
		} else {
			errcode = ptr.bindVerify(k, v)
		}
	}
	ack.WriteUInt16(errcode)
}
func Rpc_center_isvalid_bind_info2(req, ack *common.NetPack, _ common.Conn) {
	Rpc_center_isvalid_bind_info(req, ack, nil)
}
