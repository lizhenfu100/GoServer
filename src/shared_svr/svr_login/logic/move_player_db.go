/***********************************************************************
* @ 玩家数据迁移，login(http)转发
* @ brief
	1、若game是tcp节点，就无法在Rpc函数里将game结果回复给Rpc发起者
	2、需另行通知，代码逻辑散乱……很不友好

	3、TODO：针对game(tcp)，可再加个proxy(tcp)节点，统一转发game数据

* @ author zhoumf
* @ date 2019-4-3
***********************************************************************/
package logic

import (
	"common"
	"conf"
	"generate_out/err"
	"generate_out/rpc/enum"
	"netConfig"
	"netConfig/meta"
	"nets/http"
	"shared_svr/svr_login/logic/cache"
	"time"
)

func Rpc_login_move_player_db(req, ack *common.NetPack, _ common.Conn) {
	gameName := req.ReadString()
	version := req.ReadString()
	accountName := req.ReadString()
	//读取玩家数据
	accountId := req.ReadUInt32()
	playerBuf := req.ReadLenBuf()
	uid, pf_id, saveData := "", "", []byte(nil)
	if conf.HaveClientSave { //读取存档数据
		uid = req.ReadString()
		pf_id = req.ReadString()
		saveData = req.ReadLenBuf()
	}
	//3、新大区选空闲game
	if gameName != conf.GameName {
		ack.WriteUInt16(err.LoginSvr_not_match)
	} else if gameSvrId := GetFreeGameSvr(version); gameSvrId <= 0 {
		ack.WriteUInt16(err.None_free_game_server)
	} else if pGame, ok := netConfig.GetGameRpc(gameSvrId); !ok {
		ack.WriteUInt16(err.None_game_server)
	} else {
		errCode := err.Unknow_error
		defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
			ack.WriteUInt16(errCode) //同步调用，才可用ack直接回复
		}()
		//4、向game问询save，存档写入新区
		if conf.HaveClientSave {
			if e := _MoveSave(pGame, version, uid, pf_id, saveData); e != err.Success {
				errCode = e
				return
			}
		}
		//5、game创建角色，覆写（最后写角色，先写角色后写若失败，玩家能登录进新区，但数据缺损）
		if e := _MovePlayer(pGame, accountId, playerBuf); e != err.Success {
			errCode = e
			return
		}
		//6、更新center中的游戏路由
		if conf.Auto_GameSvr {
			centerId := netConfig.HashCenterID(accountName)
			netConfig.CallRpcCenter(centerId, enum.Rpc_center_set_game_route2, func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteString(conf.GameName)
				buf.WriteInt(meta.G_Local.SvrID) //loginId
				buf.WriteInt(gameSvrId)
			}, func(recvBuf *common.NetPack) {
				errCode = recvBuf.ReadUInt16()
				cache.Rpc_login_del_account_cache(recvBuf, nil, nil)
			})
		}
	}
}
func _MovePlayer(pGame common.Conn, accountId uint32, data []byte) uint16 {
	c := make(chan uint16, 1)
	pGame.CallRpc(enum.Rpc_game_move_player_db2, func(buf *common.NetPack) {
		buf.WriteUInt32(accountId)
		buf.WriteBuf(data)
	}, func(recvBuf *common.NetPack) {
		c <- recvBuf.ReadUInt16()
	})
	return _Wait(c, err.None_game_server)
}
func _MoveSave(pGame common.Conn, version, uid, pf_id string, data []byte) uint16 {
	c := make(chan uint16, 1)
	pGame.CallRpc(enum.Rpc_get_meta, func(buf *common.NetPack) {
		buf.WriteString("save")
		buf.WriteString(version)
		buf.WriteByte(common.Core)
	}, func(recvBuf *common.NetPack) {
		errCode := err.None_save_server
		if svrId := recvBuf.ReadInt(); svrId > 0 {
			ip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			http.CallRpc(http.Addr(ip, port), enum.Rpc_save_gm_up, func(buf *common.NetPack) {
				buf.WriteString(uid)
				buf.WriteString(pf_id)
				buf.WriteString("") //extra
				buf.WriteLenBuf(data)
			}, func(recvBuf *common.NetPack) {
				errCode = recvBuf.ReadUInt16()
			})
		}
		c <- errCode
	})
	return _Wait(c, err.None_game_server)
}

// ------------------------------------------------------------
// 异步转同步
func _Wait(c chan uint16, v uint16) uint16 {
	select {
	case v = <-c:
		return v
	case <-time.After(5 * time.Second):
		return v
	}
}
