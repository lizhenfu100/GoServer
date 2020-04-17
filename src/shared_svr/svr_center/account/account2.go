package account

import (
	"common"
	"common/format"
	"common/std/sign"
	"dbmgo"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/account/gameInfo"
	"sync/atomic"
	"time"
)

func Rpc_check_identity(req, ack *common.NetPack) {
	str := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString() //email、name、phone
	gameName := req.ReadString()
	gamelog.Debug("Login: %s %s", str, pwd)
	if e, p := GetAccountByBindInfo(typ, str); p == nil {
		ack.WriteUInt16(e)
	} else if !p.CheckPasswd(pwd) {
		ack.WriteUInt16(err.Account_mismatch_passwd)
	} else {
		ack.WriteUInt16(err.Success)
		timeNow := time.Now().Unix()
		atomic.StoreInt64(&p.LoginTime, timeNow)
		dbmgo.UpdateId(KDBTable, p.AccountID, bson.M{"$set": bson.M{"logintime": timeNow}})
		//1、先回复，Client账号信息
		v := gameInfo.TAccountClient{
			p.AccountID,
			p.IsValidEmail,
		}
		v.DataToBuf(ack)
		//2、再回复，附带的游戏数据，可能有的游戏空的
		if v, ok := p.GameInfo[gameName]; ok {
			ack.WriteInt(v.LoginSvrId)
			ack.WriteInt(v.GameSvrId)
		}
	}
}
func Rpc_center_account_reg2(req, ack *common.NetPack)       { Rpc_center_account_reg(req, ack) }
func Rpc_center_account_reg_force2(req, ack *common.NetPack) { Rpc_center_account_reg_force(req, ack) }
func Rpc_center_reg_if2(req, ack *common.NetPack) {
	str := req.ReadString()
	typ := req.ReadString()
	sign.Decode(&str)
	e, _ := GetAccountByBindInfo(typ, str)
	ack.WriteUInt16(e)
}
func Rpc_center_set_game_route2(req, ack *common.NetPack) { Rpc_center_set_game_route(req, ack) }
func Rpc_center_set_game_json2(req, ack *common.NetPack)  { Rpc_center_set_game_json(req, ack) }
func Rpc_center_game_info2(req, ack *common.NetPack) {
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
func Rpc_center_player_addr2(req, ack *common.NetPack) { Rpc_center_player_login_addr(req, ack) }
func Rpc_center_bind_info2(req, ack *common.NetPack) {
	str := req.ReadString()
	pwd := req.ReadString()
	typ := req.ReadString()
	v := req.ReadString()
	sign.Decode(&str, &pwd)
	errcode, ptr := GetAccountByBindInfo(typ, str)
	if ptr != nil {
		if errcode = err.Unknow_error; !ptr.CheckPasswd(pwd) {
			errcode = err.Account_mismatch_passwd
		} else if !format.CheckBindValue(typ, v) {
			errcode = err.BindInfo_format_err
		} else {
			errcode = ptr.bindVerify(typ, v)
		}
	}
	ack.WriteUInt16(errcode)
}
func Rpc_center_isvalid_bind_info2(req, ack *common.NetPack) { Rpc_center_isvalid_bind_info(req, ack) }
