package logic

import (
	"common"
	"common/assert"
	"common/std/sign"
	"common/timer"
	"common/tool/email"
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

func AccountRegLimit() {
	http.SetIntercept(func(req, ack *common.NetPack, ip string) bool {
		msgId := req.GetMsgId()
		oldPos := req.ReadPos //临时读取buffer数据
		switch msgId {
		case enum.Rpc_login_relay_to_center:
			switch rpcId := req.ReadUInt16(); rpcId { //目标rpc
			case enum.Rpc_center_account_reg:
				if _banReg(req, ack, ip) {
					return true
				}
			case enum.Rpc_center_reg_if:
				if _banRegif(ack, ip) {
					return true
				}
			default:
				if _banToCenter(ack, ip) {
					return true
				}
			}
		case enum.Rpc_login_account_login:
			if _banLogin(req, ack, ip) {
				return true
			}
		}
		req.ReadPos = oldPos
		return false
	})
}

// ------------------------------------------------------------
// -- 拦截高频调用
var (
	//10秒超5次，封5分钟
	g_regFreq    = timer.NewFreq(5, 10, 5)
	g_loginFreq  = timer.NewFreq(5, 10, 5)
	g_regifFreq  = timer.NewFreq(3, 10, 5)
	g_centerFreq = timer.NewFreq(30, 1, 5)

	G_banLogin = true
)

func _banReg(req, ack *common.NetPack, ip string) bool { //注册
	emailAddr := req.ReadString()
	passwd := req.ReadString()
	req.WriteString("email") //TODO:目前只邮箱注册

	if !assert.IsDebug && !g_regFreq.Check(ip) {
		gamelog.Info("Ban ip:%s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	if false { //要求验证邮箱
		errcode := _NeedVerifyEmail(emailAddr, passwd)
		ack.WriteUInt16(errcode)
		return true
	}
	return false
}
func _banLogin(req, ack *common.NetPack, ip string) bool { //登录
	if !assert.IsDebug && G_banLogin && !g_loginFreq.Check(ip) {
		gamelog.Info("Ban ip:%s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	return false
}
func _banRegif(ack *common.NetPack, ip string) bool {
	if !assert.IsDebug && !g_regifFreq.Check(ip) {
		gamelog.Info("Ban reg_if:%s", ip)
		ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
		return true                            //拦截，原rpc函数不会调用了
	}
	return false
}
func _banToCenter(ack *common.NetPack, ip string) bool {
	if !assert.IsDebug && !g_centerFreq.Check(ip) {
		gamelog.Info("Ban to center:%s", ip)
		ack.SetType(common.Err_too_often) //不是所有消息都返回错误码的~囧
		return true
	}
	return false
}

// ------------------------------------------------------------
// -- 邮件注册账户
func _NeedVerifyEmail(emailAddr, passwd string) (errcode uint16) {
	centerAddr := netConfig.GetHttpAddr("center", netConfig.HashCenterID(emailAddr))
	http.CallRpc(centerAddr, enum.Rpc_center_reg_check, func(buf *common.NetPack) {
		buf.WriteString(emailAddr)
		buf.WriteString(passwd)
		buf.WriteString("email")
	}, func(backBuf *common.NetPack) {
		if errcode = backBuf.ReadUInt16(); errcode == err.Success {
			//1、创建url
			u, _ := url.Parse(centerAddr + "/reg_account_by_email")
			q := u.Query()
			//2、写入参数
			q.Set("email", emailAddr)
			q.Set("pwd", passwd)
			flag := strconv.FormatInt(time.Now().Unix(), 10)
			q.Set("flag", flag)
			q.Set("language", conf.SvrCsv.EmailLanguage)
			q.Set("sign", sign.CalcSign(emailAddr+passwd+flag))
			//3、生成完整url
			u.RawQuery = q.Encode()
			errcode = email.SendMail("Verify Email", emailAddr, u.String(), "")
			if errcode == err.Success {
				errcode = err.Email_try_send_please_check
			}
		}
	})
	return
}
