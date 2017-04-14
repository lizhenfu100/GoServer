package common

const (
	HEADER_SIZE  = 3 //packetType & Opcode
	TYPE_INDEX   = 0
	OPCODE_INDEX = 1
)

type NetPack struct {
	ByteBuffer
}

func NewNetPack(capacity int) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = make([]byte, HEADER_SIZE, capacity+HEADER_SIZE)
	pack.ReadPos = HEADER_SIZE
	return pack
}
func NewNetPackLen(length int) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = make([]byte, length)
	pack.ReadPos = HEADER_SIZE
	return pack
}
func (self *NetPack) Reset(data []byte) {
	self.DataPtr = data
	self.ReadPos = HEADER_SIZE
}
func (self *NetPack) GetBody() []byte {
	return self.DataPtr[HEADER_SIZE:]
}
func (self *NetPack) SetOpCode(id uint16) {
	self.DataPtr[OPCODE_INDEX] = byte(id)
	self.DataPtr[OPCODE_INDEX+1] = byte(id >> 8)
}
func (self *NetPack) GetOpCode() uint16 {
	ret := uint16(self.DataPtr[OPCODE_INDEX+1])<<8 | uint16(self.DataPtr[OPCODE_INDEX])
	return ret
}
func (self *NetPack) SetFromTyp(typ uint8) {
	self.DataPtr[TYPE_INDEX] = typ
}
func (self *NetPack) GetFromTyp() uint8 {
	return self.DataPtr[TYPE_INDEX]
}
