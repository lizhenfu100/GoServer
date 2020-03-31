/***********************************************************************
* @ 多进程服务器架构
* @ brief
	1、shared_svr不属于某个游戏，可供其它游戏复用
		、【独立的好友服，好友关系入库】类似微信，实现社交关系不同游戏间复用

	2、唯一的Zookeeper，其它进程需在Zookeeper注册，管理后台所有进程

	3、【业务节点动态扩展】
		、游戏服，在账号上绑定区服编号（自动分服）/ 本地缓存（玩家手选区服）：确定玩家-服务器匹配关系
		、战斗服，无持久性状态数据，扩展方便；GameSvr将玩家数据通过Cross转至某Battle
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
	"conf"
	"dbmgo"
	"fmt"
	"gamelog"
	"netConfig/meta"
	"nets/http"
	http2 "nets/http/http"
	"nets/tcp"
	"sync"
)

func RunNetSvr(block bool) {
	//1、找到当前的配置信息，连db
	local := meta.G_Local
	assert.True(local != nil)
	if p := meta.GetMeta("db_"+local.Module, local.SvrID); p != nil {
		//TODO:支持连多个db，分库分表
		dbmgo.InitWithUser(p.IP, p.TcpPort, p.SvrName, conf.SvrCsv.DBuser, conf.SvrCsv.DBpasswd)
	}
	//2、连接并注册到其它模块
	if nil == meta.GetMeta("zookeeper", 0) { //没有zoo节点，才依赖配置，否则依赖zoo通知
		meta.G_Metas.Range(func(_, v interface{}) bool {
			if p := v.(*meta.Meta); p.IsMyServer(local) == meta.SC {
				ConnectModule(p)
			}
			return true
		})
	}
	//3、开启本模块网络服务(Busy Loop)
	fmt.Printf("-------%s(%d %d) start-------\n", local.Module, local.SvrID, local.Port())
	if local.HttpPort > 0 {
		http.InitSvr(local.Module, local.SvrID)
		if block {
			http2.NewHttpServer(local.HttpPort)
		} else {
			go http2.NewHttpServer(local.HttpPort)
		}
	} else if local.TcpPort > 0 {
		if block {
			tcp.NewTcpServer(local.TcpPort, local.Maxconn)
		} else {
			go tcp.NewTcpServer(local.TcpPort, local.Maxconn)
		}
	} else {
		gamelog.Error(local.Module + ": have none HttpPort|TcpPort")
	}
}

//Notice：参数pMeta会被闭包引用(且会存入容器)，须避免外界变更其内容，最好都是new的
func ConnectModule(dest *meta.Meta) {
	if dest.HttpPort > 0 {
		http.RegistToSvr(http.Addr(dest.IP, dest.HttpPort))
		meta.AddMeta(dest)
	} else if dest.TcpPort > 0 {
		ConnectModuleTcp(dest, func(*tcp.TCPConn) { meta.AddMeta(dest) })
	} else {
		gamelog.Error(dest.Module + ": have none HttpPort|TcpPort!!!")
	}
}
func ConnectModuleTcp(dest *meta.Meta, cb func(*tcp.TCPConn)) {
	if dest.TcpPort == 0 {
		gamelog.Error(dest.Module + ": have none TcpPort!!!")
		return
	}
	var client *tcp.TCPClient
	if v, ok := g_clients.Load(std.KeyPair{dest.Module, dest.SvrID}); ok {
		client = v.(*tcp.TCPClient)
	} else {
		client = new(tcp.TCPClient)
		g_clients.Store(std.KeyPair{dest.Module, dest.SvrID}, client)
	}
	client.ConnectToSvr(tcp.Addr(dest.IP, dest.TcpPort), cb)
}

var g_clients sync.Map //<{module,svrId}, *tcp.TCPClient> //本模块主动连其它模块的tcp
// TCPConn是对真正net.Conn的包装，发生断线重连时，会执行tcp.TCPConn.ResetConn()，所以外部缓存的tcp.TCPConn仍有效，无需更新
func GetTcpConn(module string, svrId int) *tcp.TCPConn {
	if v, ok := g_clients.Load(std.KeyPair{module, svrId}); ok {
		return v.(*tcp.TCPClient).Conn
	}
	// cross(s) - game(c)
	return tcp.FindRegModule(module, svrId)
}
func GetHttpAddr(module string, svrId int) string {
	if p := meta.GetMeta(module, svrId); p != nil && p.HttpPort > 0 {
		return http.Addr(p.IP, p.HttpPort)
	}
	gamelog.Debug("GetHttpAddr nil : (%s,%d)", module, svrId)
	return ""
}
