package logic

import (
	"common"
	"common/assert"
	"common/timer"
	"generate_out/err"
	"generate_out/rpc/enum"
	"nets/http"
	"sync"
	"time"
)

var g_regFreq sync.Map //<string, *timer.OpFreq>

func AccountRegLimit() { //限制同ip账号注册频率
	if !assert.IsDebug {
		http.G_Intercept = func(req, ack *common.NetPack, ip string) bool {
			msgId := req.GetOpCode()
			switch msgId {
			case enum.Rpc_login_relay_to_center:
				{
					//临时读取buffer数据
					oldPos := req.ReadPos
					rpcId := req.ReadUInt16() //目标rpc
					req.ReadPos = oldPos

					if rpcId == enum.Rpc_center_account_reg {
						freq, _ := g_regFreq.Load(ip)
						if freq == nil {
							freq = timer.NewOpFreq(10, 3600) //一小时10次
							g_regFreq.Store(ip, freq)
						}
						if !freq.(*timer.OpFreq).Check(time.Now().Unix()) {
							timer.G_TimerMgr.AddTimerSec(func() {
								g_regFreq.Delete(ip)
							}, 24*3600, 0, 0)
							ack.WriteUInt16(err.Operate_too_often) //Notice：回复内容须与原rpc一致
							return true                            //拦截，原rpc函数不会调用了
						}
					}
				}
			}
			return false
		}
	}
}
