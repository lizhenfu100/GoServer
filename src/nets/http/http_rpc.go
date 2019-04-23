package http

import (
	"common"
	"common/std/compress"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"io"
	"svr_client/test/qps"
)

var (
	G_HandleFunc [enum.RpcEnumCnt]func(req, ack *common.NetPack)
	G_Intercept  func(req, ack *common.NetPack, clientIp string) bool
)

func CallRpc(addr string, rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(64)
	req.SetOpCode(rid)
	sendFun(req)
	if buf := Client.PostReq(addr+"/client_rpc", req.Data()); buf != nil && recvFun != nil {
		if ack := common.NewNetPack(compress.Decompress(buf)); ack != nil {
			recvFun(ack)
		}
	}
	req.Free()
}
func HandleRpc(request []byte, w io.Writer, clientIp string) { //G_Intercept==nil,外界不必取ip,传空即可
	req := common.NewNetPack(request)
	if req == nil {
		gamelog.Error("invalid req: %v", request)
		return
	}
	ack := common.NewNetPackCap(128)
	defer func() {
		compress.CompressTo(ack.Data(), w)
		ack.Free()
		req.Free()
	}()
	msgId := req.GetOpCode()
	if msgId >= enum.RpcEnumCnt {
		gamelog.Error("Msg(%d) Not Regist", msgId)
		return
	}
	gamelog.Debug("HttpMsg:%d, len:%d", msgId, req.Size())

	if conf.TestFlag_CalcQPS {
		qps.AddQps()
	}
	if G_Intercept != nil { //拦截器
		if G_Intercept(req, ack, clientIp) {
			gamelog.Info("Intercept msg:%d ip:%s", msgId, clientIp)
			return
		}
	}
	if handler := G_HandleFunc[msgId]; handler != nil {
		handler(req, ack)
	} else {
		gamelog.Error("Msg(%d) Not Regist", msgId)
	}
}

// ------------------------------------------------------------
//! player rpc
type PlayerRpc struct {
	url       string
	AccountId uint32
}

var G_RecvHttpSvrData func(*common.NetPack) //对应于http_to_client.go

func NewPlayerRpc(addr string, accountId uint32) *PlayerRpc {
	return &PlayerRpc{addr + "/player_rpc", accountId}
}
func (self *PlayerRpc) CallRpc(rid uint16, sendFun, recvFun func(*common.NetPack)) {
	req := common.NewNetPackCap(64)
	req.SetOpCode(rid)
	req.SetReqIdx(self.AccountId)
	sendFun(req)
	if buf := Client.PostReq(self.url, req.Data()); buf != nil {
		if ack := common.NewNetPack(compress.Decompress(buf)); ack != nil {
			if recvFun != nil {
				recvFun(ack)
			}
			if G_RecvHttpSvrData != nil {
				G_RecvHttpSvrData(ack) //服务器主动下发的数据
			}
		}
	}
	req.Free()
}
