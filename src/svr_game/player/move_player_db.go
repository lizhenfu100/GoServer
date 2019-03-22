package player

import (
	"common"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"nets/http"
	"netConfig"
	"netConfig/meta"
)

func Rpc_game_move_player_db(req, ack *common.NetPack, this *TPlayer) {
	newLoginId := req.ReadInt()

	errCode := err.Unknow_error
	defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
		ack.WriteUInt16(errCode)
	}()

	//本节点的login
	loginAddr := ""
	if ids := meta.GetModuleIDs("login", meta.G_Local.Version); len(ids) > 0 {
		loginAddr = netConfig.GetHttpAddr("login", ids[0])
	}

	//1、向center查询新大区地址
	newLoginAddr := ""
	http.CallRpc(loginAddr, enum.Rpc_login_relay_to_center, func(buf *common.NetPack) {
		buf.WriteUInt16(enum.Rpc_meta_list)
		buf.WriteString("login")
		buf.WriteString(meta.G_Local.Version)
	}, func(recvBuf *common.NetPack) {
		cnt := recvBuf.ReadByte()
		for i := byte(0); i < cnt; i++ {
			id := recvBuf.ReadInt()
			ip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			recvBuf.ReadString() //svrName
			if id == newLoginId {
				newLoginAddr = http.Addr(ip, port)
				return
			}
		}
	})
	if newLoginAddr == "" {
		errCode = err.Svr_not_working
		return
	}

	//3、角色数据，转发至新大区
	http.CallRpc(newLoginAddr, enum.Rpc_login_move_player_db, func(buf *common.NetPack) {
		buf.WriteString(conf.GameName)
		buf.WriteString(meta.G_Local.Version)
		//game中的玩家数据迁移
		buf.WriteUInt32(this.AccountID)
		buf.WriteString(this.Name)
		playerBuf, _ := common.T2B(this)
		buf.WriteLenBuf(playerBuf)
	}, func(recvBuf *common.NetPack) {
		errCode = recvBuf.ReadUInt16()
	})

	//4、新大区选取空闲svr_game，创建角色

	//5、向game问询save地址，存档写入新区

	//6、更改center中的longSvrid、gameSvrid

	//7、迁移完成前，玩家不应向新区写数据，有脏写风险（“玩家上传存档”先于“后台节点传来的”，那玩家的新存档会被旧的覆盖掉~）
}

func Rpc_game_move_player_db2(req, ack *common.NetPack, this *TPlayer) {
	playerBuf := req.LeftBuf()

	if e := common.B2T(playerBuf, this); e == nil {
		this.WriteAllToDB()
	} else {
		gamelog.Error(e.Error())
	}
}
