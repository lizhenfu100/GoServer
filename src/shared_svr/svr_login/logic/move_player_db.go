package logic

import (
	"common"
	"common/assert"
	"common/std/sign"
	"conf"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"http"
	"netConfig"
	"netConfig/meta"
	"shared_svr/svr_center/gameInfo"
	"strconv"
)

func Rpc_login_move_player_db(req, ack *common.NetPack) {
	gameName := req.ReadString()
	version := req.ReadString()
	//读取玩家数据
	accountId := req.ReadUInt32()
	name := req.ReadString()
	playerBuf := req.ReadLenBuf()
	pf_id, mac, saveData := "", "", []byte{}
	if conf.HaveCllientSave { //读取存档数据
		pf_id = req.ReadString()
		mac = req.ReadString()
		saveData = req.ReadLenBuf()
	}

	if gameSvrId := GetFreeGameSvr(version); gameSvrId <= 0 {
		ack.WriteUInt16(err.None_free_game_server)
	} else {
		//Notice：同步调用，才可用ack直接回复 zhoumf
		errCode, isSyncCall := err.Unknow_error, false
		defer func() { //defer ack.WriteUInt16(errCode) Bug：声明时参数立即解析
			ack.WriteUInt16(errCode)
		}()
		//4、新大区选取空闲svr_game，创建角色
		gameSvrId %= 10000
		gameAddr := netConfig.GetHttpAddr("game", gameSvrId)
		http.CallRpc(gameAddr, enum.Rpc_game_create_player, func(buf *common.NetPack) {
			buf.WriteUInt32(accountId)
			buf.WriteString(name)
		}, func(recvBuf *common.NetPack) {
			isSyncCall = true
			if pid := recvBuf.ReadUInt32(); pid > 0 { //创建成功，设置玩家数据
				http.NewPlayerRpc(gameAddr, accountId).CallRpc(enum.Rpc_game_move_player_db2,
					func(buf *common.NetPack) {
						buf.WriteBuf(playerBuf)
					}, nil)
			}
		})
		//5、向game问询save地址，存档写入新区
		isSaveMoveOK := false
		if conf.HaveCllientSave {
			http.CallRpc(gameAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("save")
				buf.WriteString(version)
			}, func(recvBuf *common.NetPack) {
				isSyncCall = true
				if cnt := recvBuf.ReadByte(); cnt <= 0 {
					errCode = err.None_save_server
				} else {
					recvBuf.ReadInt() //svrId
					ip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //svrName
					saveAddr, uid := http.Addr(ip, port), strconv.Itoa(int(accountId))
					http.CallRpc(saveAddr, enum.Rpc_save_upload_binary2, func(buf *common.NetPack) {
						buf.WriteString(uid)
						buf.WriteString(pf_id)
						buf.WriteString(mac)
						buf.WriteString(sign.CalcSign(fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)))
						buf.WriteString("")
						buf.WriteLenBuf(saveData)
					}, func(recvBuf *common.NetPack) {
						if e := recvBuf.ReadUInt16(); e != err.Success {
							errCode = e
						} else {
							isSaveMoveOK = true

						}
					})
				}
			})
		} else {
			isSaveMoveOK = true
		}
		//6、更新center中的游戏信息
		if isSaveMoveOK {
			netConfig.CallRpcCenter(1, enum.Rpc_center_set_game_info,
				func(buf *common.NetPack) {
					buf.WriteUInt32(accountId)
					buf.WriteString(gameName)
					info := gameInfo.TGameInfo{
						LoginSvrId: meta.G_Local.SvrID,
						GameSvrId:  gameSvrId,
					}
					info.DataToBuf(buf)
				}, func(recvBuf *common.NetPack) {
					errCode = recvBuf.ReadUInt16()
				})
		}
		assert.True(isSyncCall)
	}
}
