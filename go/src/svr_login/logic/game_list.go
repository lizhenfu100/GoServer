package logic

import (
	"common"
	"svr_login/api"
)

func Rpc_login_get_gamesvr_lst(req, ack *common.NetPack) {
	cfgLst := api.GetRegGamesvrCfgLst()
	ack.WriteByte(byte(len(cfgLst)))
	for _, v := range cfgLst {
		ack.WriteUInt32(uint32(v.SvrID))
		ack.WriteString(v.SvrName)
		ack.WriteString(v.OutIP)
		if v.HttpPort > 0 {
			ack.WriteUInt16(uint16(v.HttpPort))
		} else {
			ack.WriteUInt16(uint16(v.TcpPort))
		}
	}
}
