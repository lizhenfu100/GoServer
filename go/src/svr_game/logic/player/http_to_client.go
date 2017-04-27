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
)

const (
	Bit_Mail_Lst  = 0
	Bit_Chat_Info = 1
)

func PackSendData(ptr interface{}, buf *common.ByteBuffer) {
	player := ptr.(*TPlayer)
	//! 先写位标记
	bit, bitPosInBuf := uint32(0), uint32(buf.Size())
	buf.WriteUInt32(bit)

	//! 再写数据块
	if pos := player.Mail.GetNoSendMailIdx(); pos >= 0 {
		player.Mail.MailLstToBuf(buf, pos)
		common.SetBit32(&bit, Bit_Mail_Lst, true)
		fmt.Println("PackSendData", bit)
	}

	//! 最后重置位标记
	_ResetBitInByteBuffer(buf, bitPosInBuf, bit)
}
func _ResetBitInByteBuffer(buf *common.ByteBuffer, pos, v uint32) {
	for i := uint32(0); i < 4; i++ {
		buf.DataPtr[pos+i] = byte(v << i)
	}
}
