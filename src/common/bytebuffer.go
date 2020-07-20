package common

import (
	"common/assert"
	"math"
	"sync"
)

type ByteBuffer struct {
	buf     []byte
	ReadPos int
}

func NewByteBufferCap(capacity int) *ByteBuffer {
	self := malloc()
	if cap(self.buf) < capacity {
		self.buf = make([]byte, 0, capacity)
	}
	return self
}
func NewByteBufferLen(length int) *ByteBuffer {
	self := malloc()
	if cap(self.buf) < length {
		self.buf = make([]byte, length)
	} else {
		self.buf = self.buf[:length]
	}
	return self
}
func ToBuf(data []byte) *ByteBuffer { return &ByteBuffer{buf: data} }

func (self *ByteBuffer) Data() []byte    { return self.buf }
func (self *ByteBuffer) Size() int       { return len(self.buf) }
func (self *ByteBuffer) LeftBuf() []byte { return self.buf[self.ReadPos:] }
func (self *ByteBuffer) Clear() {
	self.buf = self.buf[:0]
	self.ReadPos = 0
}
func (self *ByteBuffer) Reset(buf []byte, pos int) {
	self.buf = buf
	self.ReadPos = pos
}

//! Write
func (self *ByteBuffer) WriteByte(v byte)   { self.buf = append(self.buf, v) }
func (self *ByteBuffer) WriteInt(v int)     { self.WriteInt32(int32(v)) }
func (self *ByteBuffer) WriteInt8(v int8)   { self.buf = append(self.buf, byte(v)) }
func (self *ByteBuffer) WriteUInt8(v uint8) { self.buf = append(self.buf, byte(v)) }
func (self *ByteBuffer) WriteBool(v bool) {
	vv := byte(0)
	if v {
		vv = 1
	}
	self.buf = append(self.buf, vv)
}
func (self *ByteBuffer) WriteInt16(v int16) {
	self.buf = append(self.buf, byte(v), byte(v>>8))
}
func (self *ByteBuffer) WriteUInt16(v uint16) {
	self.buf = append(self.buf, byte(v), byte(v>>8))
}
func (self *ByteBuffer) WriteInt32(v int32) {
	self.buf = append(self.buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
func (self *ByteBuffer) WriteUInt32(v uint32) {
	self.buf = append(self.buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
func (self *ByteBuffer) WriteInt64(v int64) {
	self.buf = append(self.buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}
func (self *ByteBuffer) WriteUInt64(v uint64) {
	self.buf = append(self.buf, byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
}
func (self *ByteBuffer) WriteFloat(v float32) {
	self.WriteUInt32(math.Float32bits(v))
}
func (self *ByteBuffer) WriteString(v string) {
	buf := S2B(v)
	self.WriteUInt16(uint16(len(buf)))
	self.WriteBuf(buf)
}
func (self *ByteBuffer) WriteBuf(v []byte) { self.buf = append(self.buf, v...) }
func (self *ByteBuffer) WriteLenBuf(v []byte) {
	self.WriteUInt32(uint32(len(v)))
	self.buf = append(self.buf, v...)
}
func (self *ByteBuffer) ReadLenBuf() (ret []byte) {
	length := int(self.ReadUInt32())
	if self.readableBytes() >= length {
		old := self.ReadPos
		self.ReadPos += length
		ret = self.buf[old:self.ReadPos]
	}
	return
}

//! Read
func (self *ByteBuffer) readableBytes() int { //剩余多少字节没读
	return len(self.buf) - self.ReadPos
}
func (self *ByteBuffer) ReadString() (ret string) {
	length := int(self.ReadUInt16())
	if self.readableBytes() >= length {
		old := self.ReadPos
		self.ReadPos += length
		//ret = B2S(self.buf[old:self.ReadPos]) Bug:须拷贝出去，强转引用的同片内存，数据错乱
		ret = string(self.buf[old:self.ReadPos])
	}
	return
}
func (self *ByteBuffer) ReadBool() bool { return self.ReadUInt8() > 0 }
func (self *ByteBuffer) ReadByte() byte { return self.ReadUInt8() }
func (self *ByteBuffer) ReadInt() int   { return int(self.ReadInt32()) }
func (self *ByteBuffer) ReadInt8() (ret int8) {
	if self.readableBytes() >= 1 {
		ret = int8(self.buf[self.ReadPos])
		self.ReadPos += 1
	}
	return
}
func (self *ByteBuffer) ReadUInt8() (ret uint8) {
	if self.readableBytes() >= 1 {
		ret = uint8(self.buf[self.ReadPos])
		self.ReadPos += 1
	}
	return
}
func (self *ByteBuffer) ReadInt16() (ret int16) {
	if self.readableBytes() >= 2 {
		ret = int16(self.buf[self.ReadPos+1])<<8 | int16(self.buf[self.ReadPos])
		self.ReadPos += 2
	}
	return
}
func (self *ByteBuffer) ReadUInt16() (ret uint16) {
	if self.readableBytes() >= 2 {
		ret = uint16(self.buf[self.ReadPos+1])<<8 | uint16(self.buf[self.ReadPos])
		self.ReadPos += 2
	}
	return
}
func (self *ByteBuffer) ReadInt32() (ret int32) {
	if self.readableBytes() >= 4 {
		for i := 0; i < 4; i++ {
			ret |= int32(self.buf[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 4
	}
	return
}
func (self *ByteBuffer) ReadUInt32() (ret uint32) {
	if self.readableBytes() >= 4 {
		for i := 0; i < 4; i++ {
			ret |= uint32(self.buf[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 4
	}
	return
}
func (self *ByteBuffer) ReadInt64() (ret int64) {
	if self.readableBytes() >= 8 {
		for i := 0; i < 8; i++ {
			ret |= int64(self.buf[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 8
	}
	return
}
func (self *ByteBuffer) ReadUInt64() (ret uint64) {
	if self.readableBytes() >= 8 {
		for i := 0; i < 8; i++ {
			ret |= uint64(self.buf[self.ReadPos+i]) << uint(i*8)
		}
		self.ReadPos += 8
	}
	return
}
func (self *ByteBuffer) ReadFloat() (ret float32) {
	bits := self.ReadUInt32()
	return math.Float32frombits(bits)
}

//! Set
func (self *ByteBuffer) SetUInt32(pos int, v uint32) {
	if len(self.buf) >= pos+4 {
		self.buf[pos] = byte(v)
		self.buf[pos+1] = byte(v >> 8)
		self.buf[pos+2] = byte(v >> 16)
		self.buf[pos+3] = byte(v >> 24)
	}
}
func (self *ByteBuffer) GetUInt32(pos int) (ret uint32) {
	if len(self.buf) >= pos+4 {
		for i := 0; i < 4; i++ {
			ret |= uint32(self.buf[pos+i]) << uint(i*8)
		}
	}
	return
}
func (self *ByteBuffer) SetUInt16(pos int, v uint16) {
	if len(self.buf) >= pos+2 {
		self.buf[pos] = byte(v)
		self.buf[pos+1] = byte(v >> 8)
	}
}
func (self *ByteBuffer) GetUInt16(pos int) (ret uint16) {
	if len(self.buf) >= pos+2 {
		for i := 0; i < 2; i++ {
			ret |= uint16(self.buf[pos+i]) << uint(i*8)
		}
	}
	return
}

// ------------------------------------------------------------
//! 对象池
var g_pool = sync.Pool{New: func() interface{} { return new(ByteBuffer) }}

func malloc() *ByteBuffer {
	buf := g_pool.Get().(*ByteBuffer)
	buf.ReadPos = 0
	return buf
}
func (p *ByteBuffer) Free() {
	assert.True(p.ReadPos >= 0) //防重复归还
	p.buf = p.buf[:0]
	p.ReadPos = -9999 //防仍被使用
	g_pool.Put(p)
}
