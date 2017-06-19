package common

const (
	PACK_HEADER_SIZE = 7 //packetType & Opcode & reqIdx
	TYPE_INDEX       = 0 //uint8
	OPCODE_INDEX     = 1 //uint16
	REQ_IDX_INDEX    = 3 //uint32
)

type NetPack struct {
	ByteBuffer
}
type TRpcCsv struct {
	Name     string
	ID       int
	IsClient int //是否Client实现的rpc
}

var G_RpcCsv map[string]*TRpcCsv

func NewNetPackCap(capacity int) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = make([]byte, PACK_HEADER_SIZE, capacity+PACK_HEADER_SIZE)
	pack.ReadPos = PACK_HEADER_SIZE
	return pack
}
func NewNetPackLen(length int) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = make([]byte, length+PACK_HEADER_SIZE)
	pack.ReadPos = PACK_HEADER_SIZE
	return pack
}
func NewNetPack(data []byte) *NetPack {
	pack := new(NetPack)
	pack.DataPtr = data
	pack.ReadPos = PACK_HEADER_SIZE
	return pack
}
func (self *NetPack) Reset(data []byte) {
	self.DataPtr = data
	self.ReadPos = PACK_HEADER_SIZE
}
func (self *NetPack) Body() []byte  { return self.DataPtr[PACK_HEADER_SIZE:] }
func (self *NetPack) BodySize() int { return len(self.DataPtr) - PACK_HEADER_SIZE }
func (self *NetPack) ClearBody() {
	self.DataPtr = self.DataPtr[:PACK_HEADER_SIZE]
	self.ReadPos = PACK_HEADER_SIZE
}
func (self *NetPack) ResetHead(other *NetPack) {
	self.DataPtr = self.DataPtr[:0]
	self.WriteBuf(other.DataPtr[:PACK_HEADER_SIZE])
	self.ReadPos = PACK_HEADER_SIZE
}
func (self *NetPack) GetReqKey() uint64 {
	return uint64(self.GetOpCode())<<32 | uint64(self.GetReqIdx())
}

//! head
func (self *NetPack) SetOpCode(id uint16) {
	self.DataPtr[OPCODE_INDEX] = byte(id)
	self.DataPtr[OPCODE_INDEX+1] = byte(id >> 8)
}
func (self *NetPack) GetOpCode() uint16 {
	return uint16(self.DataPtr[OPCODE_INDEX+1])<<8 | uint16(self.DataPtr[OPCODE_INDEX])
}
func (self *NetPack) SetFromTyp(typ uint8) {
	self.DataPtr[TYPE_INDEX] = typ
}
func (self *NetPack) GetFromTyp() uint8 {
	return self.DataPtr[TYPE_INDEX]
}
func (self *NetPack) SetReqIdx(idx uint32) {
	self.DataPtr[REQ_IDX_INDEX] = byte(idx)
	self.DataPtr[REQ_IDX_INDEX+1] = byte(idx >> 8)
	self.DataPtr[REQ_IDX_INDEX+2] = byte(idx >> 16)
	self.DataPtr[REQ_IDX_INDEX+3] = byte(idx >> 24)
}
func (self *NetPack) GetReqIdx() (ret uint32) {
	for i := 0; i < 4; i++ {
		ret |= uint32(self.DataPtr[REQ_IDX_INDEX+i]) << uint(i*8)
	}
	return
}

//! Set
func (self *NetPack) SetPos(pos int, v uint32) { self.ByteBuffer.SetPos(PACK_HEADER_SIZE+pos, v) }
func (self *NetPack) GetPos(pos int) uint32    { return self.ByteBuffer.GetPos(PACK_HEADER_SIZE + pos) }

//! rpc
func DebugRpcIdToName(id uint16) string {
	for _, v := range G_RpcCsv {
		if v.ID == int(id) {
			return v.Name
		}
	}
	println("!!! msgId:", id, " isn't in rpc.csv  !!!\n")
	return "nil"
}
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
