/***********************************************************************
* @ 玩家数据迁移
* @ login(http)转发
	· long(http)，与它的交互是同步的
	· 若game也是http节点，转发代码就都是同步的，写起来很方便
	· game(tcp)、long(http)无法实现连续rpc链式调用……不够友好

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
	accountName := req.ReadString()

	errCode := err.Unknow_error
	defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
		ack.WriteUInt16(errCode)
	}()

	//1、向center查询新大区地址
	newLoginAddr := ""
	if p, ok := netConfig.GetRpcRand(conf.GameName); ok {
		p.CallRpcSafe(enum.Rpc_login_to_center_by_str, func(buf *common.NetPack) {
			buf.WriteUInt16(enum.Rpc_get_meta)
			buf.WriteString(conf.GameName)
			buf.WriteString(meta.G_Local.Version)
			buf.WriteByte(common.ById)
			buf.WriteInt(newLoginId)
		}, func(recvBuf *common.NetPack) {
			if svrId := recvBuf.ReadInt(); svrId > 0 {
				ip := recvBuf.ReadString()
				port := recvBuf.ReadUInt16()
				newLoginAddr = http.Addr(ip, port)
			}
		})
	}
	if newLoginAddr == "" {
		errCode = err.Svr_not_working
		return
	}
	//2、角色数据，转发至新大区
	http.CallRpc(newLoginAddr, enum.Rpc_login_move_player_db, func(buf *common.NetPack) {
		buf.WriteString(conf.GameName)
		buf.WriteString(meta.G_Local.Version)
		buf.WriteString(accountName)
		//game中的玩家数据迁移
		buf.WriteUInt32(this.AccountID)
		playerBuf, _ := common.T2B(this)
		buf.WriteLenBuf(playerBuf)
	}, func(recvBuf *common.NetPack) {
		errCode = recvBuf.ReadUInt16()
	})

	//3、新大区选空闲game

	//4、向game问询save，存档写入新区

	//5、game创建角色，覆写（最后写角色，先写角色后写若失败，玩家能登录进新区，但数据缺损）

	//6、更新center中的longSvrid、gameSvrid

	//7、迁移完成前，玩家不应向新区写数据，有脏写风险（“玩家上传存档”先于“后台节点传来的”，那玩家的新存档会被旧的覆盖掉~）
}

func Rpc_game_move_player_db2(req, ack *common.NetPack, conn *tcp.TCPConn) {
	accountId := req.ReadUInt32()
	playerBuf := req.LeftBuf()

	this := FindWithDB(accountId)
	if this == nil { //载入失败，须新建角色
		if this = NewPlayerInDB(accountId); this == nil {
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
