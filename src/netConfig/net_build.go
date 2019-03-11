/***********************************************************************
* @ 多进程服务器架构
* @ brief
	1、shared_svr不属于某个游戏，可供其它游戏复用
		、【独立的好友服，好友关系入库】类似微信，实现社交关系不同游戏间复用

	2、唯一的Zookeeper，其它进程需在Zookeeper注册，管理后台所有进程

	3、【业务节点动态扩展】
		、游戏服，在账号上绑定区服编号（自动分服）/ 本地缓存（玩家手选区服）：确定玩家-服务器匹配关系
		、战斗服，无持久性状态数据，扩展方便；GameSvr将玩家数据通过Cross转至某Battle（hash取模路由）
		、其它服务节点，如邮件、好友...设计成无状态的(redis缓存)，彼此一致，直接hash取模即可

	4、【gatew采用hash取模方式路由玩家，无法动态扩展，部署上需预留盈余，性能不够时可将同机器上其它节点迁移】

* @ reconnect
	1、【1-1】关系中的"client"重启：game每次均会连接battle
	2、【1-1】关系中的"server"重启：cross(tcp)重启，game的TCPClient.connectRoutine能检查到失败，循环重连
	3、【1-N】关系中的"N"重启：game每次均会去sdk注册
	4、【1-N】关系中的"1"重启：http_server.go会本地存储注册地址，重启时载入

* @ author zhoumf
* @ date 2016-8-11
***********************************************************************/
package netConfig

import (
	"common/assert"
	"common/std"
	"fmt"
	"gamelog"
	"http"
	"netConfig/meta"
	"sync"
	"tcp"
)

var g_client_conns sync.Map //<{module,svrId}, *tcp.TCPClient> //本模块主动连其它模块的tcp

func RunNetSvr() {
	//1、找到当前的配置信息
	assert.True(meta.G_Local != nil)

	//2、连接并注册到其它模块
	if nil == meta.GetMeta("zookeeper", 0) { //没有zookeeper节点，才依赖配置，否则依赖zookeeper的通知
		for _, connModule := range meta.G_Local.ConnectLst {
			meta.G_Metas.Range(func(k, v interface{}) bool {
				dest := v.(*meta.Meta)
				if dest.Module == connModule && !dest.IsSame(meta.G_Local) {
					ConnectModule(dest)
				}
				return true
			})
		}
	}

	//3、开启本模块网络服务(Busy Loop)
	fmt.Printf("-------%s(%d) server start-------\n", meta.G_Local.Module, meta.G_Local.SvrID)
	if meta.G_Local.HttpPort > 0 {
		http.NewHttpServer(meta.G_Local.HttpPort, meta.G_Local.Module, meta.G_Local.SvrID)
	} else if meta.G_Local.TcpPort > 0 {
		tcp.NewTcpServer(meta.G_Local.TcpPort, meta.G_Local.Maxconn)
	} else {
		gamelog.Error("%s: have none HttpPort|TcpPort!!!", meta.G_Local.Module)
	}
}

//Notice：参数pMeta会被闭包引用(且会存入容器)，须避免外界变更其内容，最好都是new的
func ConnectModule(dest *meta.Meta) {
	if dest.HttpPort > 0 {
		http.RegistToSvr(http.Addr(dest.IP, dest.HttpPort), meta.G_Local)
		meta.AddMeta(dest)
	} else if dest.TcpPort > 0 {
		ConnectModuleTcp(dest, func(*tcp.TCPConn) { meta.AddMeta(dest) })
	} else {
		gamelog.Error("%s: have none HttpPort|TcpPort!!!", dest.Module)
	}
}
func ConnectModuleTcp(dest *meta.Meta, cb func(*tcp.TCPConn)) {
	if dest.TcpPort == 0 {
		gamelog.Error("%s: have none TcpPort!!!", dest.Module)
		return
	}
	var client *tcp.TCPClient
	if v, ok := g_client_conns.Load(std.KeyPair{dest.Module, dest.SvrID}); ok {
		client = v.(*tcp.TCPClient)
	} else {
		client = new(tcp.TCPClient)
		g_client_conns.Store(std.KeyPair{dest.Module, dest.SvrID}, client)
		//Notice：client.ConnectToSvr是异步过程，这里的client.TcpConn还是空指针，不能保存*TCPConn
	}
	client.OnConnect = cb
	client.ConnectToSvr(tcp.Addr(dest.IP, dest.TcpPort), meta.G_Local)
}

// TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
func GetTcpConn(module string, svrId int) *tcp.TCPConn {
	if v, ok := g_client_conns.Load(std.KeyPair{module, svrId}); ok {
		return v.(*tcp.TCPClient).Conn
	}
	// cross(s) - game(c)
	return tcp.FindRegModule(module, svrId)
}
func GetHttpAddr(module string, svrId int) string {
	if pMeta := meta.GetMeta(module, svrId); pMeta != nil {
		return http.Addr(pMeta.IP, pMeta.HttpPort)
	}
	gamelog.Debug("GetHttpAddr nil : (%s,%d)", module, svrId)
	return ""
}
