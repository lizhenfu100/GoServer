package logic

import (
	"common"
	"common/timer"
	"conf"
	"generate_out/err"
	"generate_out/rpc/enum"
	mhttp "http"
	"net/http"
	"sync"
	"time"
)

var g_regFreq sync.Map //make(map[string]*timer.OpFreq, 5000)

func AccountRegLimit() { //限制同ip账号注册频率
	if !conf.IsDebug {
		mhttp.G_Intercept = func(req, ack *common.NetPack, ip string) bool {
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
							time.AfterFunc(72*time.Hour, func() {
								g_regFreq.Delete(ip)
							})
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

func Http_permit_ip(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ip := q.Get("ip")

	if q.Get("passwd") != conf.GM_Passwd {
		w.Write(common.S2B("passwd error"))
		return
	}
	g_regFreq.Delete(ip)
	w.Write(common.S2B("ok"))
}
