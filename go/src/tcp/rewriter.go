package tcp

import (
	"io"
)

type rewriter struct {
	data   []byte
	head   int
	length int
}

func (self *rewriter) Init(bufsize int) { self.data = make([]byte, bufsize) }

// Conn.Write() --> self.rewriter.Push(b); self.writeCount += uint64(len(b))
func (self *rewriter) Push(b []byte) {
	if len(b) >= len(self.data) {
		drop := len(b) - len(self.data)
		copy(self.data, b[drop:])
		self.head, self.length = 0, len(self.data)
		return
	}
	size := copy(self.data[self.head:], b)

	if len(b) == size {
		self.head += size
		if self.head == len(self.data) {
			self.head = 0
		}
		if self.length != len(self.data) {
			self.length += size
		}
	} else {
		self.head = copy(self.data, b[size:])
		if self.length != len(self.data) {
			self.length = len(self.data)
		}
	}
}

// Conn.doReconn() --> go self.rewriter.Rewrite()
func (self *rewriter) Rewrite(w io.Writer, writeCount, readCount uint64) bool {
	n := int(writeCount - readCount)
	switch {
	case n == 0:
		return true
	case n < 0 || n > self.length:
		return false
	case n <= self.head:
		_, err := w.Write(self.data[self.head-n : self.head])
		return err == nil
	}
	// 消息包，一半在尾一半在头
	offset := self.head - n + len(self.data)
	if _, err := w.Write(self.data[offset:]); err != nil {
		return false
	}
	_, err := w.Write(self.data[:self.head])
	return err == nil
}
