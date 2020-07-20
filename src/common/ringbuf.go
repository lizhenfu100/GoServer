package common

type RingBuf struct {
	buf []byte //长度须为2的幂
	w   uint32 //只会累加，越界回环
	r   uint32
}

func (b *RingBuf) Init(size uint32) {
	b.buf = make([]byte, CeilToPowerOfTwo(size))
	b.w, b.r = 0, 0
}
func (b *RingBuf) Read(p []byte) (n int) {
	if b.w == b.r {
		return 0
	}
	head := b.w & uint32(len(b.buf)-1) //取余后的头尾
	tail := b.r & uint32(len(b.buf)-1)
	if head > tail {
		n = copy(p, b.buf[tail:head])
	} else {
		if n = copy(p, b.buf[tail:]); n < len(p) {
			n += copy(p[n:], b.buf[:head])
		}
	}
	b.r += uint32(n)
	return
}
func (b *RingBuf) Write(p []byte) {
	cnt := uint32(len(p))
	if left := b.WritableBytes(); cnt > left {
		b.grow(cnt - left)
	}
	head := b.w & uint32(len(b.buf)-1) //取余后的头尾
	if n := copy(b.buf[head:], p); n < len(p) {
		copy(b.buf, p[n:])
	}
	b.w += cnt
}
func (b *RingBuf) Peek() (ret1 []byte, ret2 []byte) {
	if b.w == b.r {
		return
	}
	head := b.w & uint32(len(b.buf)-1) //取余后的头尾
	tail := b.r & uint32(len(b.buf)-1)
	if head > tail {
		ret1 = b.buf[tail:head]
	} else {
		ret1 = b.buf[tail:]
		ret2 = b.buf[:head]
	}
	return
}
func (b *RingBuf) ReadableBytes() uint32 {
	if b.w == b.r {
		return 0
	}
	kLen := uint32(len(b.buf))
	head := b.w & (kLen - 1) //取余后的头尾
	tail := b.r & (kLen - 1)
	if head > tail {
		return head - tail
	}
	return kLen - tail + head
}
func (b *RingBuf) WritableBytes() uint32 {
	kLen := uint32(len(b.buf))
	if b.w == b.r {
		return kLen
	}
	head := b.w & (kLen - 1) //取余后的头尾
	tail := b.r & (kLen - 1)
	if head > tail {
		return kLen - head + tail
	} else {
		return tail - head
	}
}
func (b *RingBuf) grow(n uint32) {
	newLen := CeilToPowerOfTwo(uint32(len(b.buf)) + n)
	newBuf := make([]byte, newLen)
	oldLen := b.ReadableBytes()
	b.Read(newBuf)
	b.buf = newBuf
	b.r, b.w = 0, oldLen
}
func CeilToPowerOfTwo(n uint32) uint32 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

func (b *RingBuf) Cap() uint32   { return uint32(len(b.buf)) }
func (b *RingBuf) IsFull() bool  { return b.r+uint32(len(b.buf)) == b.w }
func (b *RingBuf) IsEmpty() bool { return b.r == b.w }
func (b *RingBuf) Clear()        { b.r, b.w = 0, 0 }
