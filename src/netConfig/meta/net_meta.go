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
	1、G_Metas sync.Map 若存指针，须要求外界放入的指针，必须是堆上的，且指向不同内存
	2、最好每次存入，都重新new

* @ 动态更新
    1、运维通知zookeeper，让其将某个节点meta设置成关闭（还需同步给关联节点们）
	2、各个节点有自己的关闭策略，如：
		· 玩家相关的节点，待所有玩家下线后，可自杀
		· 转发性质的，没有逻辑状态数据的，等三分钟后自杀

* @ 多线程下，裸读写int32，除了脏读，可能读出来错数据不，比如写了几个字节，被另外线程读了？？
	· 通常都是对齐的，是不会一半被修改的
	· 不对齐的非原子读取，我持怀疑态度
	· 只要不跨cacheline, 可以保证读这4个是原子的
	· ok，那就是说，如果业务只要求最终一致，裸读，是不是就没问题？
	· 现在的程序一般都对齐了，只要送进了struct，再读，应该都对齐着的
	· 流数据不一定

* @ author zhoumf
* @ date 2017-11-30
***********************************************************************/
package meta

import (
	"common"
	"common/std"
	"gamelog"
	"math/rand"
	"sort"
	"sync"
)

var (
	G_Metas sync.Map //<{module,svrId}, pMeta>
	//G_Metas []Meta //Notice：对象数组，外界持有指针后，删除导致内存移动，所指内容改变
	G_Local *Meta
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
	Maxconn    int32
	ConnectLst []string //待连接的模块名
	Closed     bool     //TODO:如何检测节点失效？通信失败或超时
}
type Metas []Meta

func (p *Meta) Port() uint16 {
	if p.HttpPort > 0 {
		return p.HttpPort
	} else {
		return p.TcpPort
	}
}
func (v Metas) Init() {
	for i := 0; i < len(v); i++ {
		AddMeta(&v[i])
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
	buf.WriteInt32(self.Maxconn)
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
	self.Maxconn = buf.ReadInt32()
	self.ConnectLst = self.ConnectLst[:0]
	for cnt, i := buf.ReadByte(), byte(0); i < cnt; i++ {
		str := buf.ReadString()
		self.ConnectLst = append(self.ConnectLst, str)
	}
}

// -------------------------------------
//Notice：ptr必须是堆上的，且指向不同内存
func AddMeta(pNew *Meta) {
	key := std.KeyPair{pNew.Module, pNew.SvrID}
	G_Metas.Store(key, pNew)
	//if v, ok := G_Metas.Load(key); ok {
	//	*v.(*Meta) = *pNew //Bug：非线程安全的
	//} else {
	//	G_Metas.Store(key, pNew)
	//}
}
func DelMeta(module string, svrID int) {
	gamelog.Debug("DelMeta: %s:%d", module, svrID)
	key := std.KeyPair{module, svrID}
	if v, ok := G_Metas.Load(key); ok && !v.(*Meta).Closed {
		v.(*Meta).Closed = true
		G_Metas.Delete(key)
	}
}

type metas []*Meta

func (p metas) Len() int           { return len(p) }
func (p metas) Less(i, j int) bool { return p[i].SvrID < p[j].SvrID }
func (p metas) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

//Notice：禁止缓存指针，G_Metas会被多线程改写
func GetMeta(module string, svrID int) *Meta {
	if v, ok := G_Metas.Load(std.KeyPair{module, svrID}); ok && !v.(*Meta).Closed {
		return v.(*Meta)
	}
	gamelog.Debug("{%s %d}: none meta", module, svrID)
	return nil
}
func GetMetas(module, version string) (ret []*Meta) {
	G_Metas.Range(func(_, v interface{}) bool {
		if p := v.(*Meta); p.Module == module && !p.Closed &&
			common.IsMatchVersion(p.Version, version) {
			ret = append(ret, p)
		}
		return true
	})
	sort.Sort(metas(ret))
	return
}
func GetByRand(module string) *Meta {
	ret := GetMetas(module, G_Local.Version)
	if n := len(ret); n > 0 {
		return ret[rand.Intn(n)]
	}
	return nil
}
func GetByMod(module string, key uint32) *Meta {
	ret := GetMetas(module, G_Local.Version)
	if n := uint32(len(ret)); n > 0 {
		return ret[key%n]
	}
	return nil
}

//分流节点【svr_game类带状态的，须保证玩家分配到的节点不变，不能动态增删】
func ShuntSvr(svrId *int, all []*Meta, key uint32) *Meta {
	var ret []*Meta
	for _, p := range all {
		if p.SvrID%common.KIdMod == *svrId%common.KIdMod {
			ret = append(ret, p)
		}
	}
	if n := uint32(len(ret)); n > 0 {
		ptr := ret[key%n]
		*svrId = ptr.SvrID
		return ptr
	}
	return nil
}

// -------------------------------------
const (
	None = iota
	CS
	SC
	KSavePort = 7090 //固定云存档端口，便于随游戏服动态扩增
	Zookeeper = "zookeeper"
)

func (src *Meta) IsMyServer(dst *Meta) byte {
	if !dst.IsSame(src) {
		if std.Strings(src.ConnectLst).Index(dst.Module) >= 0 {
			if src.TcpPort > 0 && dst.HttpPort > 0 {
				return SC //tcp连http，改为http连tcp（保障双向通信）
			} else {
				return CS //src客户端，dst服务器
			}
		} else if std.Strings(dst.ConnectLst).Index(src.Module) >= 0 {
			if dst.TcpPort > 0 && src.HttpPort > 0 {
				return CS //tcp连http，改为http连tcp（保障双向通信）
			} else {
				return SC //src服务器，dst客户端
			}
		}
	}
	return None
}
func (src *Meta) IsSame(dst *Meta) bool { return src.Module == dst.Module && src.SvrID == dst.SvrID }
