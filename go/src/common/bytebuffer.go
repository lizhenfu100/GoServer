package common

import (
	"math"
)

type ByteBuffer struct {
	DataPtr []byte
	ReadPos int
}

func NewByteBufferCap(capacity int) *ByteBuffer {
	buf := new(ByteBuffer)
	buf.DataPtr = make([]byte, 0, capacity)
	return buf
}
func NewByteBufferLen(length int) *ByteBuffer {
	buf := new(ByteBuffer)
	buf.DataPtr = make([]byte, length)
	return buf
}
func NewByteBuffer(data []byte) *ByteBuffer { return &ByteBuffer{data, 0} }
func (self *ByteBuffer) Reset(data []byte) {
	self.DataPtr = data
	self.ReadPos = 0
}
func (self *ByteBuffer) Clear() {
	self.DataPtr = self.DataPtr[:0]
	self.ReadPos = 0
}
func (self *ByteBuffer) Size() int { return len(self.DataPtr) }

//! Write
func (self *ByteBuffer) WriteByte(v byte) {
	self.DataPtr = append(self.DataPtr, v)
}
func (self *ByteBuffer) WriteInt(v int) { self.WriteInt32(int32(v)) }
func (self *ByteBuffer) WriteInt8(v int8) {
	self.DataPtr = append(self.DataPtr, byte(v))
}
func (self *ByteBuffer) WriteUInt8(v uint8) {
	self.DataPtr = append(self.DataPtr, byte(v))
}
func (self *ByteBuffer) WriteInt16(v int16) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8))
}
func (self *ByteBuffer) WriteUInt16(v uint16) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8))
}
func (self *ByteBuffer) WriteInt32(v int32) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
func (self *ByteBuffer) WriteUInt32(v uint32) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
func (self *ByteBuffer) WriteInt64(v int64) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}
func (self *ByteBuffer) WriteUInt64(v uint64) {
	self.DataPtr = append(self.DataPtr, byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}
func (self *ByteBuffer) WriteFloat(v float32) {
	self.WriteUInt32(math.Float32bits(v))
}
func (self *ByteBuffer) WriteString(v string) {
	bytes := []byte(v)
	self.WriteUInt16(uint16(len(bytes)))
	self.WriteBuf(bytes)
}
func (self *ByteBuffer) WriteBuf(v []byte) {
	self.DataPtr = append(self.DataPtr, v...)
}

//! Read
func (self *ByteBuffer) readableBytes() int { //剩余多少字节没读
	return len(self.DataPtr) - self.ReadPos
}
func (self *ByteBuffer) ReadFloat() (ret float32) {
	bits := self.ReadUInt32()
	return math.Float32frombits(bits)
}
func (self *ByteBuffer) ReadString() (ret string) {
	length := int(self.ReadUInt16())
	if self.readableBytes() >= length {
		bytes := self.DataPtr[self.ReadPos : self.ReadPos+length]
		self.ReadPos += length
		ret = string(bytes)
	}
	return
}
func (self *ByteBuffer) ReadByte() (ret byte) {
	if self.readableBytes() >= 1 {
		ret = self.DataPtr[self.ReadPos]
		self.ReadPos += 1
	}
	return
}
func (self *ByteBuffer) ReadInt() int { return int(self.ReadInt32()) }
func (self *ByteBuffer) ReadInt8() (ret int8) {
	if self.readableBytes() >= 1 {
		ret = int8(self.DataPtr[self.ReadPos])
		self.ReadPos += 1
	}
	return
}
func (self *ByteBuffer) ReadUInt8() (ret uint8) {
	if self.readableBytes() >= 1 {
		ret = uint8(self.DataPtr[self.ReadPos])
		self.ReadPos += 1
	}
	return
}
func (self *ByteBuffer) ReadInt16() (ret int16) {
	if self.readableBytes() >= 2 {
		ret = int16(self.DataPtr[self.ReadPos+1])<<8 | int16(self.DataPtr[self.ReadPos])
		self.ReadPos += 2
	}
	return
}
func (self *ByteBuffer) ReadUInt16() (ret uint16) {
	if self.readableBytes() >= 2 {
		ret = uint16(self.DataPtr[self.ReadPos+1])<<8 | uint16(self.DataPtr[self.ReadPos])
		self.ReadPos += 2
	}
	return
}
func (self *ByteBuffer) ReadInt32() (ret int32) {
	if self.readableBytes() >= 4 {
		for i := 0; i < 4; i++ {
			ret |= int32(self.DataPtr[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 4
	}
	return
}
func (self *ByteBuffer) ReadUInt32() (ret uint32) {
	if self.readableBytes() >= 4 {
		for i := 0; i < 4; i++ {
			ret |= uint32(self.DataPtr[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 4
	}
	return
}
func (self *ByteBuffer) ReadInt64() (ret int64) {
	if self.readableBytes() >= 8 {
		for i := 0; i < 8; i++ {
			ret |= int64(self.DataPtr[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 8
	}
	return
}
func (self *ByteBuffer) ReadUInt64() (ret uint64) {
	if self.readableBytes() >= 8 {
		for i := 0; i < 8; i++ {
			ret |= uint64(self.DataPtr[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 8
	}
	return
}

//! Set
func (self *ByteBuffer) SetPos(pos int, v uint32) {
	if len(self.DataPtr) >= pos+4 {
		self.DataPtr[pos] = byte(v)
		self.DataPtr[pos+1] = byte(v >> 8)
		self.DataPtr[pos+2] = byte(v >> 16)
		self.DataPtr[pos+3] = byte(v >> 24)
	}
}
func (self *ByteBuffer) GetPos(pos int) (ret uint32) {
	if len(self.DataPtr) >= pos+4 {
		for i := 0; i < 4; i++ {
			ret |= uint32(self.DataPtr[pos+i]) << uint(i*8)
		}
	}
	return
}
