package logic

import (
	"common"
	"svr_login/api"
)

func Rpc_login_get_gamesvr_lst(req, ack *common.NetPack) {
	api.WriteRegGamesvr(ack)
}
