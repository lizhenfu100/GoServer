package tcp

import (
	"sync"
)

type TcpConnMgr struct {
	sync.Mutex
	ConnMap map[int32]*TCPConn
}

func NewConnMgr() *TcpConnMgr {
	mgr := new(TcpConnMgr)
	mgr.ConnMap = make(map[int32]*TCPConn)
	return mgr
}
func (self *TcpConnMgr) AddConnByID(id int32, pTcpConn *TCPConn) {
	self.Lock()
	self.ConnMap[id] = pTcpConn
	self.Unlock()
}
func (self *TcpConnMgr) GetConnByID(id int32) *TCPConn {
	self.Lock()
	pConn, _ := self.ConnMap[id]
	self.Unlock()
	return pConn
}
func (self *TcpConnMgr) DelConnByID(id int32) {
	self.Lock()
	delete(self.ConnMap, id)
	self.Unlock()
}
