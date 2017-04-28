/***********************************************************************
* @ http
* @ brief

* @ 通信技巧
	1、客户端20秒轮询一次服务端，问服务端有没有什么消息给我，比如双人聊天消息。
	2、如果取到消息，就把下一次轮训时间改短，比如5秒，再取到消息，继续改短，比如2秒，
	3、如果没消息就慢慢放长周期，比如：2秒，3秒，5秒，7秒，10秒，15秒，20秒
	4、直到有消息了，又再次把周期变短
	5、聊天模块的缩短程度，可以单独做大些

* @ author zhoumf
* @ date 2017-4-25
***********************************************************************/
package http

import (
	"bytes"
	"common"
	"gamelog"
	"net/http"
	"time"
)

//////////////////////////////////////////////////////////////////////
//! rpc
type ClientRpc struct {
	Url      string
	PlayerId uint32
}

func NewClientRpc(addr string, pid uint32) *ClientRpc {
	return &ClientRpc{addr + "client_rpc", pid}
}
func (self *ClientRpc) CallRpc(rpc string, sendFun, recvFun func(*common.NetPack)) {
	buf := common.NewNetPackCap(64)
	buf.SetRpc(rpc)
	buf.SetReqIdx(self.PlayerId)
	sendFun(buf)
	b := PostReq(self.Url, buf.DataPtr)
	recvFun(common.NewNetPack(b))
}

//////////////////////////////////////////////////////////////////////
//! 底层接口，业务层一般用不到
func PostMsg(url string, pMsg interface{}) []byte {
	b, _ := common.ToBytes(pMsg)
	return PostReq(url, b)
}
func PostReq(url string, b []byte) []byte {
	ack, err := http.Post(url, "text/HTML", bytes.NewReader(b))
	if err == nil {
		backBuf := make([]byte, ack.ContentLength)
		ack.Body.Read(backBuf)
		ack.Body.Close()
		return backBuf
	} else {
		gamelog.Error3("PostReq url: %s \r\nerr: %s \r\n", url, err.Error())
		return nil
	}
}

//! 模块注册
type Msg_Regist_To_HttpSvr struct {
	Addr   string
	Module string
	ID     int
}

func RegistToSvr(destAddr, srcAddr, srcModule string, srcID int) {
	go _RegistToSvr(destAddr, srcAddr, srcModule, srcID)
}
func _RegistToSvr(destAddr, srcAddr, srcModule string, srcID int) {
	pMsg := &Msg_Regist_To_HttpSvr{srcAddr, srcModule, srcID}
	for {
		http.DefaultClient.Timeout = 2 * time.Second
		if PostMsg(destAddr+"reg_to_svr", pMsg) == nil {
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}
}
