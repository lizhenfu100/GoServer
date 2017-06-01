package tcp

import (
	"io"
)

type rereader struct {
	head  *rereadNode
	tail  *rereadNode
	count uint64
}
type rereadNode struct {
	data []byte
	next *rereadNode
}

// Conn.Read() --> self.rereader.Pull(b)
func (self *rereader) Pull(b []byte) (n int) {
	if self.head != nil {
		copy(b, self.head.data)
		if len(self.head.data) > len(b) {
			self.head.data = self.head.data[len(b):]
			n = len(b)
		} else {
			n = len(self.head.data)
			self.head = self.head.next
			if self.head == nil {
				self.tail = nil
			}
		}
	}
	self.count -= uint64(n)
	return
}

// Conn.doReconn() --> go self.rereader.Reread()
func (self *rereader) Reread(rd io.Reader, n int) bool {
	b := make([]byte, n)
	if _, err := io.ReadFull(rd, b); err != nil {
		return false
	}
	node := &rereadNode{b, nil}
	if self.head == nil {
		self.head = node
	} else {
		self.tail.next = node
	}
	self.tail = node
	self.count += uint64(n)
	return true
}
