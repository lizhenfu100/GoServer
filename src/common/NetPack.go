package common

const (
	PACK_HEADER_SIZE = 7 //type & msgId & reqIdx
	INDEX_TYPE       = 0 //uint8
	INDEX_MSG_ID     = 1 //uint16
	INDEX_REQ_IDX    = 3 //uint32

	//Type == 0 默认是对端的请求，其它值是对端的回复；两边有同样rpc时区分
	Type_ack = 1
	// INDEX_TYPE：写通用错误码
	Err_offline   = 255
	Err_too_often = 254
	Err_flag      = 200
)

type NetPack = ByteBuffer

type Conn interface {
	CallRpc(msgId uint16, sendFun, recvFun func(*NetPack))
	WriteMsg(msg *NetPack)
	GetUser() interface{}
	SetUser(v interface{})
	IsClose() bool
	Close()
}

func NewNetPackCap(capacity int) *NetPack {
	self := NewByteBufferCap(capacity)
	self.ReadPos = PACK_HEADER_SIZE        //len == 0
	self.buf = self.buf[:PACK_HEADER_SIZE] //len == PACK_HEADER_SIZE
	return self
}
func ToNetPack(data []byte) *NetPack {
	if len(data) < PACK_HEADER_SIZE {
		return nil
	} else {
		return &NetPack{data, PACK_HEADER_SIZE}
	}
}

func (self *NetPack) Body() []byte  { return self.buf[PACK_HEADER_SIZE:] }
func (self *NetPack) BodySize() int { return len(self.buf) - PACK_HEADER_SIZE }
func (self *NetPack) ClearBody() {
	self.buf = self.buf[:PACK_HEADER_SIZE]
	self.ReadPos = PACK_HEADER_SIZE
}
func (self *NetPack) ResetHead(other *NetPack) {
	self.buf = self.buf[:PACK_HEADER_SIZE]
	copy(self.buf, other.buf)
	self.ReadPos = PACK_HEADER_SIZE
}

//! head
func (self *NetPack) SetType(v uint8) { self.buf[INDEX_TYPE] = v }
func (self *NetPack) GetType() uint8  { return self.buf[INDEX_TYPE] }
func (self *NetPack) SetMsgId(id uint16) {
	self.buf[INDEX_MSG_ID] = byte(id)
	self.buf[INDEX_MSG_ID+1] = byte(id >> 8)
}
func (self *NetPack) GetMsgId() uint16 {
	return uint16(self.buf[INDEX_MSG_ID+1])<<8 | uint16(self.buf[INDEX_MSG_ID])
}
func (self *NetPack) SetReqIdx(idx uint32) {
	self.buf[INDEX_REQ_IDX] = byte(idx)
	self.buf[INDEX_REQ_IDX+1] = byte(idx >> 8)
	self.buf[INDEX_REQ_IDX+2] = byte(idx >> 16)
	self.buf[INDEX_REQ_IDX+3] = byte(idx >> 24)
}
func (self *NetPack) GetReqIdx() (ret uint32) {
	for i := 0; i < 4; i++ {
		ret |= uint32(self.buf[INDEX_REQ_IDX+i]) << uint(i*8)
	}
	return
}
func (self *NetPack) GetReqKey() uint64 {
	return uint64(self.GetMsgId())<<32 | uint64(self.GetReqIdx())
}
func (self *NetPack) SetReqKey(key uint64) {
	self.SetMsgId(uint16(key >> 32))
	self.SetReqIdx(uint32(key))
}
