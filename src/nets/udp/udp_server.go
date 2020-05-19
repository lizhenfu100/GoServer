package udp

import (
	"net"
)

type UDPServer struct {
	UDPConn
	MaxConnNum int32
	connCnt    int32
}

var _svr UDPServer

func NewServer(port int, maxconn int32, block bool) {
	_svr.MaxConnNum = maxconn
	if c, e := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port}); e == nil {
		if _svr.conn = c; block {
			_svr.readLoop()
		} else {
			go _svr.readLoop()
		}
	} else {
		panic("NewServer: %s" + e.Error())
	}
}
func CloseServer() { _svr.Close() }
