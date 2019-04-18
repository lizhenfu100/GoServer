/***********************************************************************
* @ 玩家数据迁移
* @ login(http)转发
	· long(http)，与它的交互是同步的
	· 若game也是http节点，转发代码就都是同步的，写起来很方便
	· game(tcp)、long(http)无法实现连续rpc链式调用……不够友好

* @ proxy(tcp)转发
	· TODO：针对game(tcp)，新增一个proxy(tcp)，统一转发game数据

* @ author zhoumf
* @ date 2019-4-3
***********************************************************************/
package player

import (
	"common"
	"conf"
	"gamelog"
	"generate_out/err"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
)

func Rpc_game_move_player_db(req, ack *common.NetPack, this *TPlayer) {
	newLoginId := req.ReadInt()

	errCode := err.Unknow_error
	defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
		ack.WriteUInt16(errCode)
	}()

	//本节点的login
	loginAddr := netConfig.GetLoginAddr()

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

func Rpc_game_move_player_db2(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	playerName := req.ReadString()
	playerBuf := req.LeftBuf()

	this := FindWithDB(accountId)
	if this == nil { //载入失败，须新建角色
		if this = NewPlayerInDB(accountId, playerName); this == nil {
			ack.WriteUInt16(err.Create_Player_failed)
			return
		}
	}
	if e := common.B2T(playerBuf, this); e == nil {
		this.WriteAllToDB()
		ack.WriteUInt16(err.Success)
	} else {
		ack.WriteUInt16(err.Convert_err)
		gamelog.Error(e.Error())
	}
}
