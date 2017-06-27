/***********************************************************************
* @ http: svr to client
* @ brief
	1、http svr，不能主动发数据给client
	2、抽出一个单独模块，每次client请求上来，捎带数据下去
	3、可优化为红点提示，然后client打开界面时再请求对应数据

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

const (
	Bit_Mail_Lst      = 0
	Bit_Chat_Info     = 1
	Bit_Friend_Apply  = 2
	Bit_Invite_Friend = 3
	Bit_Team_Update   = 4
	Bit_Show_UI_Wait  = 5
	Bit_Team_Chat     = 6
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
	self := ptr.(*TPlayer)
	pid := self.PlayerID
	bit, bitPosInBody := uint32(0), buf.BodySize()
	//! 先写位标记
	buf.WriteUInt32(bit)
	//! 再写数据块
	if pos := self.Mail.GetNoSendIdx(); pos >= 0 {
		common.SetBit32(&bit, Bit_Mail_Lst, true)
		//self.Mail.DataToBuf(buf, pos)
		//界面红点提示
	}
	if pos := self.Chat.GetNoSendIdx(); pos >= 0 {
		common.SetBit32(&bit, Bit_Chat_Info, true)
		//界面红点提示
	}
	if len(self.Friend.ApplyLst) > 0 {
		common.SetBit32(&bit, Bit_Friend_Apply, true)
		length := len(self.Friend.ApplyLst)
		buf.WriteByte(byte(length))
		for i := 0; i < length; i++ {
			self.Friend.ApplyLst[i].DataToBuf(buf)
		}
	}
	if self.Friend.inviteMsg.Size() > 0 { //被别人邀请
		common.SetBit32(&bit, Bit_Invite_Friend, true)
		buf.WriteBuf(self.Friend.inviteMsg.DataPtr)
		self.Friend.inviteMsg.Clear()
	}
	if self.pTeam != nil && self.pTeam.isChange {
		common.SetBit32(&bit, Bit_Team_Update, true)
		self.pTeam.isChange = false
	}
	if self.Battle.isShowWaitUI { //队长开始匹配，队员得到通知
		common.SetBit32(&bit, Bit_Show_UI_Wait, true)
		self.Battle.isShowWaitUI = false
	}
	if self.pTeam != nil {
		if pos := self.pTeam.GetNoSendIdx(pid); pos >= 0 {
			common.SetBit32(&bit, Bit_Team_Chat, true)
			self.pTeam.DataToBuf(buf, pid)
		}
	}
	//! 最后重置位标记
	buf.SetPos(bitPosInBody, bit)
	if bit > 0 {
		fmt.Println("pid:", pid, "PackSendBit", bit, buf)
	}
}
