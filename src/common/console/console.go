package console

import (
	"common"
	"common/file"
	"common/std/hash"
	"common/tool/wechat"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"math/rand"
	"netConfig/meta"
	"nets/http"
	"nets/tcp"
	"os"
	"time"
)

func Init() {
	rand.Seed(time.Now().UnixNano())
	tcp.G_HandleFunc[enum.Rpc_log] = _Rpc_log1
	http.G_HandleFunc[enum.Rpc_log] = _Rpc_log2
	tcp.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_gm_cmd1
	http.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_gm_cmd2
	tcp.G_HandleFunc[enum.Rpc_meta_list] = _Rpc_meta_list1
	http.G_HandleFunc[enum.Rpc_meta_list] = _Rpc_meta_list2
	tcp.G_HandleFunc[enum.Rpc_get_meta] = _Rpc_get_meta1
	http.G_HandleFunc[enum.Rpc_get_meta] = _Rpc_get_meta2
	tcp.G_HandleFunc[enum.Rpc_update_file] = _Rpc_update_file1
	http.G_HandleFunc[enum.Rpc_update_file] = _Rpc_update_file2
	tcp.G_HandleFunc[enum.Rpc_reload_csv] = _Rpc_reload_csv1
	http.G_HandleFunc[enum.Rpc_reload_csv] = _Rpc_reload_csv2
	tcp.G_HandleFunc[enum.Rpc_timestamp] = _Rpc_timestamp1
	http.G_HandleFunc[enum.Rpc_timestamp] = _Rpc_timestamp2
	go sigTerm() //监控进程终止信号
	wechat.Init( //启动微信报警
		conf.SvrCsv.WechatCorpId,
		conf.SvrCsv.WechatSecret,
		conf.SvrCsv.WechatAgentId)
}

func _Rpc_log1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_log2(req, ack) }
func _Rpc_log2(req, ack *common.NetPack) {
	log := req.ReadString()
	uuid := req.ReadString()
	version := req.ReadString()
	pf_id := req.ReadString()
	gamelog.Info("%s, UUID:%s (%s, %s)", log, uuid, version, pf_id)
}

func _Rpc_meta_list1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_meta_list2(req, ack) }
func _Rpc_meta_list2(req, ack *common.NetPack) {
	module := req.ReadString() //game、save、file...
	version := req.ReadString()
	list := meta.GetMetas(module, version)
	ack.WriteByte(byte(len(list)))
	for _, p := range list {
		ack.WriteInt(p.SvrID)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())
		ack.WriteString(p.SvrName)
	}
}
func _Rpc_get_meta1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_get_meta2(req, ack) }
func _Rpc_get_meta2(req, ack *common.NetPack) {
	module := req.ReadString() //game、save、file...
	version := req.ReadString()
	typ := req.ReadByte()
	var p *meta.Meta
	vs := meta.GetMetas(module, version)
	if n := len(vs); n > 0 {
		switch typ {
		case common.Random:
			p = vs[rand.Intn(n)]
		case common.Core:
			p = vs[0]
		case common.ById:
			p = meta.GetMeta(module, req.ReadInt())
		case common.ModInt:
			p = vs[req.ReadUInt32()%uint32(n)]
		case common.ModStr:
			p = vs[hash.StrHash(req.ReadString())%uint32(n)]
		case common.KeyShuntInt:
			svrId := req.ReadInt()
			aid := req.ReadUInt32()
			p = meta.ShuntSvr(&svrId, vs, aid)
		case common.KeyShuntStr:
			svrId := req.ReadInt()
			hashId := hash.StrHash(req.ReadString())
			p = meta.ShuntSvr(&svrId, vs, hashId)
		}
	}
	if p != nil {
		ack.WriteInt(p.SvrID)
		ack.WriteString(p.OutIP)
		ack.WriteUInt16(p.Port())
	}
}
func _Rpc_timestamp1(req, ack *common.NetPack, _ *tcp.TCPConn) { _Rpc_timestamp2(req, ack) }
func _Rpc_timestamp2(req, ack *common.NetPack) {
	ack.WriteInt64(time.Now().Unix())
}

//TODO:动态加载配置文件，有多线程访问竞态，竞态木问题，写完最终一致的，关键是会不会宕机？配表里头有map~囧
/*
	Tcp可将更新、读取文件抽离，都读到内存后，于帧循环末尾一并刷新配置 …… 本质是找个StopWorld时机
	Http貌似没啥好办法 …… 拦截器？抽离文件读写，条件变量，待内存刷新完，拦截器才放开
		拦截不完全啊，拦截生效同时，可能有rpc正在执行，正执行的木办法了 …… 本质还是要找StopWorld时机
*/
func _Rpc_update_file1(req, ack *common.NetPack, _ *tcp.TCPConn) { go _Rpc_update_file2(req, ack) } //TcpRpc主线程调的，不应直接加载
func _Rpc_update_file2(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		dir := req.ReadString()
		name := req.ReadString()
		data := req.ReadLenBuf()
		if fd, e := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); e == nil {
			_, e = fd.Write(data)
			fd.Close()
			if e == nil {
				ack.WriteString("ok")
			} else {
				ack.WriteString(e.Error())
			}
		} else {
			ack.WriteString(e.Error())
		}
	}
}
func _Rpc_reload_csv1(req, ack *common.NetPack, _ *tcp.TCPConn) { go _Rpc_reload_csv2(req, ack) }
func _Rpc_reload_csv2(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		file.ReloadCsv(req.ReadString())
	}
	ack.WriteByte(1) //ok
}
