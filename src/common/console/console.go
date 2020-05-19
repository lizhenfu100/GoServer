package console

import (
	"common"
	"common/file"
	"common/std/hash"
	"common/tool/sms"
	"common/tool/wechat"
	"conf"
	"gamelog"
	"generate_out/rpc/enum"
	"math/rand"
	"netConfig/meta"
	"nets/rpc"
	"os"
	"time"
)

func Init() {
	rand.Seed(time.Now().UnixNano())
	rpc.G_HandleFunc[enum.Rpc_log] = _Rpc_log
	rpc.G_HandleFunc[enum.Rpc_gm_cmd] = _Rpc_gm_cmd
	rpc.G_HandleFunc[enum.Rpc_meta_list] = _Rpc_meta_list
	rpc.G_HandleFunc[enum.Rpc_get_meta] = _Rpc_get_meta
	rpc.G_HandleFunc[enum.Rpc_update_file] = _Rpc_update_file
	rpc.G_HandleFunc[enum.Rpc_reload_csv] = _Rpc_reload_csv
	rpc.G_HandleFunc[enum.Rpc_timestamp] = _Rpc_timestamp
	go sigTerm() //监控进程终止信号
	csv := conf.SvrCsv()
	wechat.Init( //微信报警
		csv.WechatCorpId,
		csv.WechatSecret,
		csv.WechatAgentId)
	sms.Init( //短信
		csv.SmsKeyId,
		csv.SmsSecret)
}

func _Rpc_log(req, ack *common.NetPack, _ common.Conn) {
	log := req.ReadString()
	uuid := req.ReadString()
	version := req.ReadString()
	pf_id := req.ReadString()
	gamelog.Info("%s, UUID:%s (%s, %s)", log, uuid, version, pf_id)
}

func _Rpc_meta_list(req, ack *common.NetPack, _ common.Conn) {
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
func _Rpc_get_meta(req, ack *common.NetPack, _ common.Conn) {
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
func _Rpc_timestamp(req, ack *common.NetPack, _ common.Conn) {
	ack.WriteInt64(time.Now().Unix())
}

/* TODO:动态加载配置文件，有多线程访问竞态，竞态木问题，写完最终一致的，关键是会不会宕机？配表里头有map~囧
载入文件、业务更新，分成两步；载入文件单独处理，“业务更新”只需【交换内存引用】
主逻辑单线程的架构，于帧循环末尾一并刷新配置 …… 本质是找个StopWorld时机
Http …… 拦截器？
	拦截生效时，可能有rpc正在读配置 …… 本质还是要找StopWorld时机
	c++的话，封装下配置类，让【交换内存引用】线程安全

将配置变量私有化、引用语义，get|set原子更改，业务使用的都是旧变量的引用拷贝，原始变量可安全的更改
	· 运行逻辑是：拷贝出shared_ptr给外部使用，内部shared_ptr指向新内存块即可
*/
func _Rpc_update_file(req, ack *common.NetPack, _ common.Conn) { go _Rpc_update_file1(req, ack) }
func _Rpc_update_file1(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		dir := req.ReadString()
		name := req.ReadString()
		data := req.ReadLenBuf()
		if fd, e := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); e == nil {
			_, e = fd.Write(data)
			fd.Close()
			if e == nil {
				file.ReloadCsv(dir + name)
				ack.WriteString("ok")
			} else {
				ack.WriteString(e.Error())
			}
		} else {
			ack.WriteString(e.Error())
		}
	}
}
func _Rpc_reload_csv(req, ack *common.NetPack, _ common.Conn) { go _Rpc_reload_csv1(req, ack) }
func _Rpc_reload_csv1(req, ack *common.NetPack) {
	for cnt, i := req.ReadByte(), byte(0); i < cnt; i++ {
		file.ReloadCsv(req.ReadString())
	}
	ack.WriteByte(1) //ok
}
