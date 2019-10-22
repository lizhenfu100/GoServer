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

* @ 通信技巧
	1、客户端20秒轮询一次服务端，问服务端有没有什么消息给我，比如双人聊天消息。
	2、如果取到消息，就把下一次轮训时间改短，比如5秒，再取到消息，继续改短，比如2秒，
	3、如果没消息就慢慢放长周期，比如：2秒，3秒，5秒，7秒，10秒，15秒，20秒
	4、直到有消息了，又再次把周期变短
	5、聊天模块的缩短程度，可以单独做大些

* @ author zhoumf
* @ date 2017-4-26
***********************************************************************/
package player

import (
	"common"
	"conf"
	"gamelog"
)

const (
	Bit_Mail_Lst      = 0
	Bit_Chat_Info     = 1
	Bit_Friend_Apply  = 2
	Bit_Invite_Friend = 3
	Bit_Team_Update   = 4
	Bit_Show_UI_Wait  = 5
	Bit_Team_Chat     = 6
	Bit_Friend_Update = 7
)

//! 需要主动发给玩家的数据，每回通信时捎带过去
func BeforeRecvHttpMsg(accountId uint32) *TPlayer {
	if player := FindAccountId(accountId); player != nil {
		if !player.IsOnline() {
			player.Login(nil)
		}
		//player._HandleAsyncNotify()
		//player.mail.SendSvrMailAll()
		return player
	}
	return nil
}
func AfterRecvHttpMsg(self *TPlayer, buf *common.NetPack) {
	if !conf.Is_Http_To_Client {
		return
	}
	accountId := self.AccountID

	//! 先写位标记
	buf.WriteUInt8(0xFF)
	bit, bitPosInBody := uint32(0), buf.BodySize()
	buf.WriteUInt32(bit)

	//! 再写数据块
	/*
		if pos := self.mail.GetNoSendIdx(); pos >= 0 {
			common.SetBit32(&bit, Bit_Mail_Lst, true)
			//界面红点提示
		}
		if pos := self.Chat.GetNoSendIdx(); pos >= 0 {
			common.SetBit32(&bit, Bit_Chat_Info, true)
			//界面红点提示
		}
		if len(self.Friend.ApplyLst) > 0 { //收到好友申请
			common.SetBit32(&bit, Bit_Friend_Apply, true)
			self.Friend.PackApplyInfo(buf)
		}
		if self.Friend.inviteMsg.Size() > 0 { //被别人邀请
			common.SetBit32(&bit, Bit_Invite_Friend, true)
			buf.WriteBuf(self.Friend.inviteMsg.Data())
			self.Friend.inviteMsg.Clear()
		}
		if self.pTeam != nil && self.pTeam.isChange { //队伍人员变动
			common.SetBit32(&bit, Bit_Team_Update, true)
			self.pTeam.isChange = false
		}
		if self.battle.isShowWaitUI { //队长开始匹配，队员得到通知
			common.SetBit32(&bit, Bit_Show_UI_Wait, true)
			self.battle.isShowWaitUI = false
		}
		if self.pTeam != nil { //队伍聊天
			if pos := self.pTeam.GetNoSendIdx(accountId); pos >= 0 {
				common.SetBit32(&bit, Bit_Team_Chat, true)
				self.pTeam.DataToBuf(buf, accountId)
			}
		}
		if self.Friend.isChange { //好友列表变动
			common.SetBit32(&bit, Bit_Friend_Update, true)
			self.Friend.isChange = false
		}
	*/
	//! 最后重置位标记
	buf.SetPos(bitPosInBody, bit)
	if bit > 0 {
		gamelog.Debug("aid(%d), PackSendBit(%b) %v", accountId, bit, buf)
	}
}

func Rpc_game_heart_beat(req, ack *common.NetPack, this *TPlayer) {
}

// ------------------------------------------------------------
//! for other player write my data
/*
func AsyncNotifyPlayer(accountId uint32, handler func(*TPlayer)) {
	if player := FindAccountId(accountId); player != nil {
		player.AsyncNotify(handler)
	}
}
func (self *TPlayer) AsyncNotify(handler func(*TPlayer)) {
	if self.IsOnline() {
		select {
		case self.askchan <- handler:
		default:
			gamelog.Warn("Player askChan is full !!!")
			return
		}
	} else { //FIXME:zhoumf: 如何安全方便的修改离线玩家数据……应该不允许的~囧，除非特殊玩法

		//准备将离线的操作转给mainloop，这样所有离线玩家就都在一个chan里处理了
		//要是中途玩家上线，mainloop的chan里还有他的操作没处理完怎么整！？囧~
		//mainloop设计成map<accountId, chan>，玩家上线时，检测自己的chan有效否，等它处理完？

		//gen_server
		//将某个独立模块的所有操作扔进gen_server，外界只读(有滞后性)
		//会加大代码量，每个操作都得转一次到chan
		//Notice：可能gen_server里还有修改操作，且玩家已下线，会重新读到内存，此时修改完毕后须及时入库

		//设计统一的接口，编辑离线数据，也很麻烦呐
	}
}
func (self *TPlayer) _HandleAsyncNotify() {
	for {
		select {
		case handler := <-self.askchan:
			handler(self)
		default:
			return
		}
	}
}
*/
