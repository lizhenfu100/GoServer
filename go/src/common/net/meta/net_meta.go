/***********************************************************************
* @ 网络节点的元数据
* @ brief
	1、每个节点一份meta信息，这份meta可以构建自csv表，也能构建自zookeeper
	2、一个节点连上zookeeper时，将与其相关的节点meta下发

* @ Notice
	1、G_SvrNets []Meta 作为一个数组，中间元素被删除后会整体移动
	2、此时若外界缓存了 GetMeta() 返回的指针，其指向很可能变为下个元素

* @ 动态更新
    1、运维通知zookeeper，让其将某个节点meta设置成关闭（还需同步给关联节点们）
	2、各个节点有自己的关闭策略，如：
		· 玩家相关的节点，待所有玩家下线后，可自杀
		· 转发性质的，没有逻辑状态数据的，等三分钟后自杀

* @ author zhoumf
* @ date 2017-11-30
***********************************************************************/
package meta

import (
	"common"
	"gamelog"
)

type Meta struct {
	Module     string
	SvrID      int
	SvrName    string
	Version    string
	IP         string //内部局域网IP
	OutIP      string
	TcpPort    uint16
	HttpPort   uint16
	Maxconn    int
	ConnectLst []string //待连接的模块名

	//需动态同步的数据
	IsClosed bool
}

// -------------------------------------
//! buf
func (self *Meta) DataToBuf(buf *common.NetPack) {
	buf.WriteString(self.Module)
	buf.WriteInt(self.SvrID)
	buf.WriteString(self.SvrName)
	buf.WriteString(self.Version)
	buf.WriteString(self.IP)
	buf.WriteString(self.OutIP)
	buf.WriteUInt16(self.TcpPort)
	buf.WriteUInt16(self.HttpPort)
	buf.WriteInt(self.Maxconn)
	length := len(self.ConnectLst)
	buf.WriteByte(byte(length))
	for i := 0; i < length; i++ {
		buf.WriteString(self.ConnectLst[i])
	}
}
func (self *Meta) BufToData(buf *common.NetPack) {
	self.Module = buf.ReadString()
	self.SvrID = buf.ReadInt()
	self.SvrName = buf.ReadString()
	self.Version = buf.ReadString()
	self.IP = buf.ReadString()
	self.OutIP = buf.ReadString()
	self.TcpPort = buf.ReadUInt16()
	self.HttpPort = buf.ReadUInt16()
	self.Maxconn = buf.ReadInt()
	length := buf.ReadByte()
	for i := byte(0); i < length; i++ {
		str := buf.ReadString()
		self.ConnectLst = append(self.ConnectLst, str)
	}
}

// -------------------------------------
//! meta list
var G_SvrNets []Meta

//Notice：缓存Meta指针有风险，详见文件头注释
func GetMeta(module string, svrID int) *Meta { //负ID表示自动找首个
	for i := 0; i < len(G_SvrNets); i++ {
		v := &G_SvrNets[i]
		if !v.IsClosed && v.Module == module && (svrID < 0 || v.SvrID == svrID) {
			return v
		}
	}
	gamelog.Error("{%s %d}: have none SvrNetMeta", module, svrID)
	return nil
}
func AddMeta(ptr *Meta) {
	for i := 0; i < len(G_SvrNets); i++ {
		v := &G_SvrNets[i]
		if v.Module == ptr.Module && v.SvrID == ptr.SvrID {
			*v = *ptr
			return
		}
	}
	G_SvrNets = append(G_SvrNets, *ptr)
}
func DelMeta(module string, svrID int) {
	for i := len(G_SvrNets) - 1; i >= 0; i-- {
		v := &G_SvrNets[i]
		if v.Module == module && v.SvrID == svrID {
			//Notice：防止内存移动，不删元素，仅改状态
			//G_SvrNets = append(G_SvrNets[:i], G_SvrNets[i+1:]...)
			v.IsClosed = true
			return
		}
	}
}

func GetIpPort(module string, id int) (ip string, port uint16) {
	if v := GetMeta(module, id); v != nil {
		if v.HttpPort > 0 {
			port = uint16(v.HttpPort)
		} else {
			port = uint16(v.TcpPort)
		}
		ip = v.IP
	}
	return
}
func GetModuleIDs(module string) (ret []int) {
	for i := 0; i < len(G_SvrNets); i++ {
		v := &G_SvrNets[i]
		if v.Module == module {
			ret = append(ret, v.SvrID)
		}
	}
	return
}

// -------------------------------------
//! logic
func (self *Meta) IsMyServer(dst *Meta) bool {
	for _, v := range self.ConnectLst {
		if v == dst.Module {
			return true
		}
	}
	return false
}
func (self *Meta) IsMyClient(dst *Meta) bool { return dst.IsMyServer(self) }

func (self *Meta) IsSame(dst *Meta) bool {
	return self.Module == dst.Module &&
		self.SvrID == dst.SvrID &&
		self.Version == dst.Version
}
