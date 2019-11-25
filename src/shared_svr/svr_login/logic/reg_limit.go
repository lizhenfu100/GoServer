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
	"sync"
	"time"
)

func AccountRegLimit() {
	http.SetIntercept(func(req, ack *common.NetPack, ip string) bool {
		msgId := req.GetOpCode()
		oldPos := req.ReadPos //临时读取buffer数据
		switch msgId {
		case enum.Rpc_login_relay_to_center:
			{
				rpcId := req.ReadUInt16() //目标rpc

				if rpcId == enum.Rpc_center_account_reg {
					if _banReg(req, ack, ip) {
						return true
					}
				}
			}
		case enum.Rpc_login_account_login:
			{
				if _banLogin(req, ack, ip) {
					return true
				}
			}
		}
		req.ReadPos = oldPos
		return false
	})
}

// ------------------------------------------------------------
// -- 注册拦截
var g_regFreq sync.Map //<ip, *timer.OpFreq>

func _banReg(req, ack *common.NetPack, ip string) bool {
	emailAddr := req.ReadString()
	passwd := req.ReadString()
	req.WriteString("email") //TODO:目前只邮箱注册

	if !assert.IsDebug {
		freq, _ := g_regFreq.Load(ip)
		if freq == nil {
			freq = timer.NewOpFreq(5, 10) //10秒超5次
			g_regFreq.Store(ip, freq)
		}
		if !freq.(*timer.OpFreq).Check(time.Now().Unix()) {
			timer.G_TimerMgr.AddTimerSec(func() {
				g_regFreq.Delete(ip)
			}, 300, 0, 0)
			gamelog.Info("Ban ip:%s", ip)
			ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
			return true                            //拦截，原rpc函数不会调用了
		}
	}
	if false { //要求验证邮箱
		errcode := _NeedVerifyEmail(emailAddr, passwd)
		ack.WriteUInt16(errcode)
		return true
	}
	return false
}

// ------------------------------------------------------------
// -- 登录拦截
var g_loginFreq sync.Map //<ip, *timer.OpFreq>

func _banLogin(req, ack *common.NetPack, ip string) bool {
	if !assert.IsDebug {
		freq, _ := g_loginFreq.Load(ip)
		if freq == nil {
			freq = timer.NewOpFreq(5, 10) //10秒超5次
			g_loginFreq.Store(ip, freq)
		}
		if !freq.(*timer.OpFreq).Check(time.Now().Unix()) {
			timer.G_TimerMgr.AddTimerSec(func() {
				g_loginFreq.Delete(ip)
			}, 300, 0, 0)
			gamelog.Info("Ban ip:%s", ip)
			ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
			return true                            //拦截，原rpc函数不会调用了
		}
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
