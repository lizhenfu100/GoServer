package common

const (
	HEADER_SIZE  = 3 //packetType & Opcode
	TYPE_INDEX   = 0 //uint8
	OPCODE_INDEX = 1 //uint16
)

type NetPack struct {
	ByteBuffer
}
type TRpcCsv struct {
	Name     string
	ID       int
	IsClient int //是否Client实现的rpc
	Comment  string
}

var G_RpcCsv map[string]*TRpcCsv

func NewNetPackCap(capacity int) *NetPack {
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
func NewNetPack(data []byte) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = data
	pack.ReadPos = HEADER_SIZE
	return pack
}
func (self *NetPack) Reset(data []byte) {
	self.DataPtr = data
	self.ReadPos = HEADER_SIZE
}
func (self *NetPack) BodyBytes() int {
	return len(self.DataPtr) - HEADER_SIZE
}
func (self *NetPack) GetBody() []byte {
	return self.DataPtr[HEADER_SIZE:]
}
func (self *NetPack) ClearBody() {
	head := self.DataPtr[:HEADER_SIZE]
	ClearBuf(&self.DataPtr)
	self.WriteBuf(head)
	self.ReadPos = HEADER_SIZE
}

//! head
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

//! rpc
func RpcNameToId(rpc string) uint16 {
	if csv, ok := G_RpcCsv[rpc]; ok {
		return uint16(csv.ID)
	}
	print("!!!  " + rpc + " isn't in rpc.csv  !!!\n")
	return 0
}
func (self *NetPack) SetRpc(name string) {
	self.SetOpCode(RpcNameToId(name))
}
