package common

type RingBuf struct {
	buf    []byte
	size   int
	r      int
	w      int
	isFull bool
}

func NewRingBuf(size int) *RingBuf {
	return &RingBuf{
		buf:  make([]byte, size),
		size: size,
	}
}
func (b *RingBuf) Read(p []byte) (n int) {
	if b.w > b.r {
		if n = b.w - b.r; n > len(p) {
			n = len(p)
		}
		copy(p, b.buf[b.r:b.r+n])
		b.r += n
	} else if !b.IsEmpty() {
		if n = b.size - b.r + b.w; n > len(p) {
			n = len(p)
		}
		copy(p, b.buf[b.r:])

		if left := b.size - b.r; left >= n {
			b.r += n
		} else {
			copy(p[left:], b.buf[:n-left])
			b.r = n - left
		}
		b.isFull = false
	}
	return
}
func (b *RingBuf) Write(p []byte) {
	n := len(p)
	if left := b.writableBytes(); left < n {
		b.malloc(n - left)
	}
	copy(b.buf[b.w:], p)

	if left := b.size - b.w; b.w >= b.r && n > left {
		copy(b.buf, p[left:])
		b.w = n - left
	} else {
		b.w += n
	}
	if b.w == b.size {
		b.w = 0
	}
	if b.w == b.r {
		b.isFull = true
	}
}
func (b *RingBuf) Peek() (head []byte, tail []byte) {
	if b.w > b.r {
		head = b.buf[b.r:b.w]
	} else if !b.IsEmpty() {
		head = b.buf[b.r:]
		tail = b.buf[:b.w]
	}
	return
}
func (b *RingBuf) readableBytes() int {
	if b.r == b.w {
		if b.isFull {
			return b.size
		}
		return 0
	}
	if b.w > b.r {
		return b.w - b.r
	}
	return b.size - b.r + b.w
}
func (b *RingBuf) writableBytes() int {
	if b.r == b.w {
		if b.isFull {
			return 0
		}
		return b.size
	}
	if b.r > b.w {
		return b.r - b.w
	}
	return b.size - b.w + b.r
}
func (b *RingBuf) malloc(n int) {
	newLen := b.size + CeilToPowerOfTwo(n)
	newBuf := make([]byte, newLen)
	oldLen := b.readableBytes()
	b.Read(newBuf)
	b.r, b.w = 0, oldLen
	b.size = newLen
	b.buf = newBuf
}
func CeilToPowerOfTwo(n int) int {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

func (b *RingBuf) Cap() int      { return b.size }
func (b *RingBuf) IsFull() bool  { return b.isFull }
func (b *RingBuf) IsEmpty() bool { return b.r == b.w && !b.isFull }
func (b *RingBuf) Clear()        { b.r, b.w, b.isFull = 0, 0, false }
