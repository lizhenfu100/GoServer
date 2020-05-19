// +build net_multi

package rpc

import "common"

func (self *RpcQueue) Insert(conn common.Conn, req *common.NetPack) {
	if msgFunc := G_HandleFunc[req.GetMsgId()]; msgFunc != nil {
		msgFunc(req, nil, conn)
	}
}
