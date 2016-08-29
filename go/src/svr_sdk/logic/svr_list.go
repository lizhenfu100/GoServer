/***********************************************************************
* @ 内部游戏服的地址列表
* @ brief
    1、SDK进程负责接收所有第三方消息，验证后转发至对应的gamesvr

    2、须预先加载服务器列表

* @ author zhoumf
* @ date 2016-8-16
***********************************************************************/
package logic

import (
	"common"
	"encoding/json"
	"gamelog"
	"http"
	"netConfig"
	"strconv"
)

var (
	SvrAddr_PATH = common.GetCurrPath() + "svr_addr.csv"

	SvrID_Addr = make(map[int]string)
)

func LoadSvrAddrCsv() {
	records, err := common.LoadCsv(SvrAddr_PATH)
	if err != nil {
		gamelog.Error("LoadSvrAddrCsv : %s", err.Error())
		return
	}

	//首行时表头，跳过
	for i := 1; i < len(records); i++ {
		id, _ := strconv.Atoi(records[i][0])
		SvrID_Addr[id] = records[i][1]
	}
}
func UpdateSvrAddrCsv() {
	//保持配表ID顺序，全部写一遍文件
	records, i := make([][]string, len(SvrID_Addr)+1), 1
	records[0] = append(records[0], "svrID", "url")
	for k, v := range SvrID_Addr {
		records[i] = append(records[i], strconv.Itoa(k), v)
		i++
	}
	if err := common.UpdateCsv(SvrAddr_PATH, records); err != nil {
		gamelog.Error("UpdateSvrAddrCsv : %s", err.Error())
	}
}
func RegisterGamesvrAddr(svrID int, url string) {
	oldUrl, ok := SvrID_Addr[svrID]
	if ok && oldUrl == url {
		return
	} else {
		SvrID_Addr[svrID] = url
		UpdateSvrAddrCsv()
	}
}

// strKey = "/sdk_recharge_info"
func RelayToGamesvr(svrId int, strKey string, pMsg interface{}) {
	// url, ok := SvrID_Addr[svrId]
	url, ok := netConfig.GetHttpAddr("game", svrId), true
	if ok {
		url += strKey
		data, _ := json.Marshal(pMsg)
		if _, err := http.PostReq(url, data); err != nil {
			gamelog.Error("RelayToGamesvr--PostServerReq: svrId(%d) %s", svrId, err.Error())
		}
	} else {
		gamelog.Error("RelayToGamesvr: svrId(%d)", svrId)
	}
}
