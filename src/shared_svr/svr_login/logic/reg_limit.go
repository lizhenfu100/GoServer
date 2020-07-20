package logic

import (
	"common"
	"common/assert"
	"common/format"
	"common/std/sign"
	"common/timer"
	"common/tool/email"
	"common/tool/sms"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/url"
	"netConfig"
	"nets/http"
	"strconv"
	"time"
)

func AccountLimit(req, ack *common.NetPack, ip string) bool {
	msgId := req.GetMsgId()
	oldPos := req.ReadPos //临时读取buffer数据
	switch msgId {
	case enum.Rpc_login_to_center_by_str:
		switch rpcId := req.ReadUInt16(); rpcId { //目标rpc
		case enum.Rpc_center_account_reg, enum.Rpc_center_account_reg2:
			if _banReg(req, ack, ip) {
				return true
			}
		case enum.Rpc_center_reg_if, enum.Rpc_center_reg_if2, enum.Rpc_center_isvalid_bind_info:
			if _banRegif(ack, ip) {
				return true
			}
		default:
			if _banToCenter(ack, ip) {
				return true
			}
		}
	case enum.Rpc_login_account_login, enum.Rpc_check_identity:
		if _banLogin(ack, ip) {
			return true
		}
	}
	req.ReadPos = oldPos
	return false
}

// ------------------------------------------------------------
// -- 拦截高频调用
var (
	g_regFreq    = timer.NewFreq(3, 10)
	g_loginFreq  = timer.NewFreq(3, 10)
	g_regifFreq  = timer.NewFreq(3, 10)
	g_centerFreq = timer.NewFreq(20, 5)

	G_banLogin = true
)

func _banReg(req, ack *common.NetPack, ip string) bool { //注册
	if !assert.IsDebug && !g_regFreq.Check(ip) {
		gamelog.Info("Ban reg: %s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	str := req.ReadString() //渠道走的Rpc_center_platform_reg
	pwd := req.ReadString()
	if len(req.LeftBuf()) == 0 { //TODO:待删除，兼容老客户端
		req.WriteString("email")
	}
	if typ := req.ReadString(); typ == "email" {
		if false { //要求验证邮箱
			errcode := _NeedVerifyEmail(str, pwd)
			ack.WriteUInt16(errcode)
			return true
		}
	}
	return false
}
func _banLogin(ack *common.NetPack, ip string) bool { //登录
	if !assert.IsDebug && G_banLogin && !g_loginFreq.Check(ip) {
		gamelog.Info("Ban login: %s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	return false
}
func _banRegif(ack *common.NetPack, ip string) bool {
	if !assert.IsDebug && !g_regifFreq.Check(ip) {
		gamelog.Info("Ban regif: %s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	return false
}
func _banToCenter(ack *common.NetPack, ip string) bool {
	if !assert.IsDebug && !g_centerFreq.Check(ip) {
		gamelog.Info("Ban to center: %s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	return false
}

// ------------------------------------------------------------
// -- 功能开关
var _switch = map[string]*bool{
	"banLogin": &G_banLogin,
	"sms":      &sms.G_Switch,
	"email":    &email.G_Switch,
}

func FlagSwitch(args []string) string { //banLogin 0
	flag, open := args[0], args[1] == "1"
	if p, ok := _switch[flag]; ok {
		if *p = open; !open {
			timer.AddTimer(func() { *p = true }, 3600, 0, 0)
		}
		return args[0] + " " + args[1]
	} else {
		return "none: " + flag
	}
}

// ------------------------------------------------------------
// -- 邮件注册账户
func _NeedVerifyEmail(emailAddr, passwd string) (errcode uint16) {
	if sign.Decode(&passwd); !format.CheckPasswd(passwd) {
		return err.Passwd_format_err
	}
	centerAddr := netConfig.GetHttpAddr("center", netConfig.HashCenterID(emailAddr))
	http.CallRpc(centerAddr, enum.Rpc_center_reg_if2, func(buf *common.NetPack) {
		buf.WriteString(emailAddr)
		buf.WriteString("email")
	}, func(backBuf *common.NetPack) {
		if errcode = backBuf.ReadUInt16(); errcode == err.Not_found {
			//1、创建url
			u, _ := url.Parse(centerAddr + "/reg_account_by_email")
			q := u.Query()
			//2、写入参数
			q.Set("email", emailAddr)
			q.Set("pwd", passwd)
			flag := strconv.FormatInt(time.Now().Unix(), 10)
			q.Set("flag", flag)
			q.Set("language", conf.SvrCsv().EmailLanguage)
			q.Set("sign", sign.CalcSign(emailAddr+passwd+flag))
			//3、生成完整url
			u.RawQuery = q.Encode()
			errcode = email.SendMail("Verify Email", emailAddr, u.String(), "")
			if errcode == err.Success {
				errcode = err.Email_try_send_please_check
			}
		} else if errcode == err.Success {
			errcode = err.Account_repeat
		}
	})
	return
}
