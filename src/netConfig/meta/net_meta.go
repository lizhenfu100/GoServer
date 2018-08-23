/***********************************************************************
* @ 网络节点的元数据
* @ brief
	1、每个节点一份meta信息，这份meta可以构建自csv表，也能构建自zookeeper
	2、一个节点连上zookeeper时，将与其相关的节点meta下发
	3、版本号格式：1.12.233，前两组一致的版本间可匹配，第三组用于小调整、bug修复
	4、空版本号能与任意版本匹配

* @ Notice
	1、G_Metas []Meta 作为一个数组，中间元素被删除后会整体移动
	2、此时若外界缓存了 GetMeta() 返回的指针，其指向很可能变为下个元素

* @ Notice
	1、G_Metas sync.Map 存的指针，这样就要求外界放入的指针，必须是堆上的，且指向不同内存
	2、最好每次存入，都重新new

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
	"sync"
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

func (self *Meta) Port() uint16 {
	if self.HttpPort > 0 {
		return self.HttpPort
	} else {
		return self.TcpPort
	}
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
	self.ConnectLst = self.ConnectLst[:0]
	for i := byte(0); i < length; i++ {
		str := buf.ReadString()
		self.ConnectLst = append(self.ConnectLst, str)
	}
}

// -------------------------------------
//! meta list
var G_Metas sync.Map

func InitConf(list []Meta) {
	for i := 0; i < len(list); i++ {
		AddMeta(&list[i])
	}
}

func GetMeta(module string, svrID int) *Meta {
	if v, ok := G_Metas.Load(common.KeyPair{module, svrID}); ok && !v.(*Meta).IsClosed {
		return v.(*Meta)
	}
	gamelog.Error("{%s %d}: have none SvrNetMeta", module, svrID)
	return nil
}

//{
//	for i := len(G_Metas) - 1; i >= 0; i-- {
//		v := &G_Metas[i]
//		if v.Module == module && v.SvrID == svrID {
//			//Notice：防止内存移动，不删元素，仅改状态
//			//G_Metas = append(G_Metas[:i], G_Metas[i+1:]...)
//			v.IsClosed = true
//			return
//		}
//	}
//}
func DelMeta(module string, svrID int) {
	gamelog.Debug("DelMeta: %s:%d", module, svrID)
	G_Metas.Delete(common.KeyPair{module, svrID})
}
func AddMeta(ptr *Meta) {
	gamelog.Debug("AddMeta: %s:%d", ptr.Module, ptr.SvrID)
	G_Metas.Store(common.KeyPair{ptr.Module, ptr.SvrID}, ptr)
}

func GetModuleIDs(module, version string) (ret []int, ok bool) {
	G_Metas.Range(func(k, v interface{}) bool {
		ptr := v.(*Meta)
		if ptr.Module == module && !ptr.IsClosed && ptr.IsMatchVersion(version) {
			ret = append(ret, ptr.SvrID)
		}
		return true
	})
	return ret, len(ret) > 0
}

// -------------------------------------
//! logic
func (self *Meta) IsMyServer(dst *Meta) bool {
	for _, v := range self.ConnectLst {
		if v == dst.Module && !dst.IsSame(self) {
			return true
		}
	}
	return false
}
func (self *Meta) IsMyClient(dst *Meta) bool { return dst.IsMyServer(self) }

func (self *Meta) IsSame(dst *Meta) bool { return self.Module == dst.Module && self.SvrID == dst.SvrID }

func (self *Meta) IsMatchVersion(version string) bool {
	// 版本号格式：1.12.233，前两组一致的版本间可匹配，第三组用于小调整、bug修复
	// 空版本号能与任意版本匹配
	return version == "" ||
		self.Version == "" ||
		self.Version[:4] == version[:4]
}
