/***********************************************************************
* @ http: svr to client
* @ brief
	1、http svr，不能主动发数据给client
	2、抽出一个单独模块，每次client请求上来，捎带数据下去

* @ 打包格式
	1、按块区分，各块分别解析/反解析
	2、头部是位标记，标识包含那些数据块
	3、数据块须按约定顺序写，否则会错乱……这种时候知道protobuf的好了吧(▔﹁▔)~

* @ author zhoumf
* @ date 2017-4-26
***********************************************************************/
package player

import (
	"common"
	"fmt"
	"sync/atomic"
)

//TODO:zhoumf: 这里可以优化为红点提示，然后client打开界面时再请求相应模块数据
const (
	Bit_Mail_Lst      = 0
	Bit_Chat_Info     = 1
	Bit_Friend_Apply  = 2
	Bit_Invite_Friend = 3
	Bit_Team_Update   = 4
)

func BeforeRecvHttpMsg(pid uint32) interface{} {
	if player := _FindInCache(pid); player != nil {
		atomic.SwapUint32(&player.idleSec, 0)
		player._HandleAsyncNotify()
		player.Mail.SendSvrMailAll()
		return player
	}
	return nil
}
func AfterRecvHttpMsg(ptr interface{}, buf *common.NetPack) {
	player := ptr.(*TPlayer)

	bit, bitPosInBody := uint32(0), buf.BodySize()
	//! 先写位标记
	buf.WriteUInt32(bit)
	//! 再写数据块
	if pos := player.Mail.GetNoSendIdx(); pos >= 0 {
		common.SetBit32(&bit, Bit_Mail_Lst, true)
		//player.Mail.DataToBuf(buf, pos)
	}
	if pos := player.Chat.GetNoSendIdx(); pos >= 0 {
		common.SetBit32(&bit, Bit_Chat_Info, true)
		//player.Chat.DataToBuf(buf, pos)
	}
	if len(player.Friend.ApplyLst) > 0 {
		common.SetBit32(&bit, Bit_Friend_Apply, true)
		length := uint16(len(player.Friend.ApplyLst))
		buf.WriteUInt16(length)
		for i := uint16(0); i < length; i++ {
			player.Friend.ApplyLst[i].DataToBuf(buf)
		}
	}
	if player.Friend.inviteMsg.Size() > 0 { //被别人邀请
		common.SetBit32(&bit, Bit_Invite_Friend, true)
		buf.WriteBuf(player.Friend.inviteMsg.DataPtr)
		player.Friend.inviteMsg.Clear()
	}
	if player.Friend.isTeamChange {
		common.SetBit32(&bit, Bit_Team_Update, true)
		player.Friend.isTeamChange = false
	}
	//! 最后重置位标记
	buf.SetPos(bitPosInBody, bit)
	if bit > 0 {
		fmt.Println("pid:", player.PlayerID, "PackSendBit", bit, buf)
	}
}
