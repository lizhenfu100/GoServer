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
	"shared_svr/svr_center/account/gameInfo"
	"strconv"
)

func Rpc_login_move_player_db(req, ack *common.NetPack) {
	gameName := req.ReadString()
	version := req.ReadString()
	//读取玩家数据
	accountId := req.ReadUInt32()
	playerBuf := req.ReadLenBuf()
	pf_id, mac, clientVersion, saveData := "", "", "", []byte(nil)
	if conf.HaveClientSave { //读取存档数据
		pf_id = req.ReadString()
		mac = req.ReadString()
		clientVersion = req.ReadString()
		saveData = req.ReadLenBuf()
	}
	if gameName != conf.GameName {
		ack.WriteUInt16(err.LoginSvr_not_match)
	} else if gameSvrId := GetFreeGameSvr(version); gameSvrId <= 0 {
		ack.WriteUInt16(err.None_free_game_server)
	} else {
		errCode := err.Unknow_error
		defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
			ack.WriteUInt16(errCode) //同步调用，才可用ack直接回复
		}()
		//4、新大区选取空闲svr_game，创建角色
		if e := _MovePlayer(gameSvrId, accountId, playerBuf); e != err.Success {
			errCode = e
			return
		}
		//5、向game问询save地址，存档写入新区
		if conf.HaveClientSave {
			if e := _MoveSave(gameSvrId, accountId, version, pf_id, mac, clientVersion, saveData); e != err.Success {
				errCode = e
				return
			}
		}
		//6、更新center中的游戏信息
		netConfig.CallRpcCenter(1, enum.Rpc_center_set_game_info,
			func(buf *common.NetPack) {
				buf.WriteUInt32(accountId)
				buf.WriteString(conf.GameName)
				info := gameInfo.TGameInfo{
					GameSvrId:  gameSvrId % 10000,
					LoginSvrId: meta.G_Local.SvrID % 10000,
				}
				info.DataToBuf(buf)
			}, func(recvBuf *common.NetPack) {
				errCode = recvBuf.ReadUInt16()
			})
	}
}
func _MovePlayer(gameSvrId int, accountId uint32, data []byte) (errCode uint16) {
	errCode = err.None_game_server
	gameAddr := netConfig.GetHttpAddr("game", gameSvrId)
	http.CallRpc(gameAddr, enum.Rpc_game_move_player_db2, func(buf *common.NetPack) {
		buf.WriteUInt32(accountId)
		buf.WriteBuf(data)
	}, func(recvBuf *common.NetPack) {
		errCode = recvBuf.ReadUInt16()
	})
	return
}
func _MoveSave(gameSvrId int, accountId uint32, version, pf_id, mac, clientVersion string, data []byte) (errCode uint16) {
	errCode = err.None_game_server
	gameAddr := netConfig.GetHttpAddr("game", gameSvrId)
	http.CallRpc(gameAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
		buf.WriteString("save")
		buf.WriteString(version)
	}, func(recvBuf *common.NetPack) {
		errCode = err.None_save_server
		if cnt := recvBuf.ReadByte(); cnt > 0 {
			recvBuf.ReadInt() //svrId
			ip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			recvBuf.ReadString() //svrName
			saveAddr, uid := http.Addr(ip, port), strconv.Itoa(int(accountId))
			http.CallRpc(saveAddr, enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
				buf.WriteString(uid)
				buf.WriteString(pf_id)
				buf.WriteString(mac)
				buf.WriteString("") //sign
				buf.WriteString("") //extra
				buf.WriteLenBuf(data)
				buf.WriteString(clientVersion)
			}, func(recvBuf *common.NetPack) {
				errCode = recvBuf.ReadUInt16()
			})
		}
	})
	return
}
