package gm

import (
	"common"
	"common/file"
	"conf"
	"dbmgo"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"netConfig/meta"
	http2 "nets/http"
	"os"
	"path/filepath"
	conf2 "shared_svr/svr_save/conf"
	"shared_svr/svr_save/logic"
	"text/template"
)

func Http_clear_maccnt(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok && ptr.MacCnt > 0 {
		dbmgo.UpdateId(logic.KDBSave, ptr.Key, bson.M{
			"$inc": bson.M{"maccnt": -1},
			"$set": bson.M{"chtime": 0},
		})
		w.Write([]byte("maccnt_add && clear_bind_limit: ok"))
	} else {
		w.Write([]byte("none save data"))
	}
}
func Http_clear_bind_limit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	key := logic.GetSaveKey(pf_id, uid)
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", key, &logic.TSaveData{}); ok {
		dbmgo.UpdateId(logic.KDBSave, key, bson.M{"$set": bson.M{"chtime": 0}})
		w.Write([]byte("clear_bind_limit: ok"))
	} else {
		w.Write([]byte("none save data"))
	}
}
func Http_clear_unbind_limit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	logic.ClearUnbindLimit()
	w.Write([]byte("clear_unbind_limit: ok"))
}

func Http_del_save_data(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") != conf.GM_Passwd {
		w.Write([]byte("passwd error"))
		return
	}
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")

	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); ok {
		ptr.Backup()
		dbmgo.Remove(logic.KDBSave, bson.M{"_id": ptr.Key})
		dbmgo.RemoveAll(logic.KDBMac, bson.M{"key": ptr.Key})
	}
	w.Write([]byte("del_save_data: ok"))
}

// 存档回退
const kTemplate = `<html>
<body bgcolor="white">
    <table border="1" cellpadding="3" cellspacing="0">
{{range $_, $v := .Names}}
        <tr>
            <td>{{$v}}</td>
			<td>
				<form action="{{$.Addr}}" method="get"><br>
                	<input type="hidden" name="pf_id" value={{$.Pf_id}}>
                	<input type="hidden" name="uid" value={{$.Uid}}>
                	<input type="hidden" name="filename" value={{$v}}>
					<input type="submit" value="回退">
				</form>
			</td>
        </tr>
{{end}}
    </table>
</body>
</html>`

type backupAck struct {
	Names []string
	Pf_id string
	Uid   string
	Addr  string
}

func Http_show_backup_file(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")
	addr := q.Get("addr")
	key := logic.GetSaveKey(pf_id, uid)
	var files []string
	var e error
	if files, e = filepath.Glob(fmt.Sprintf("player/%s/*.save", key)); e == nil && len(files) > 0 {
		for i, v := range files {
			files[i] = filepath.Base(v)
		}
	} else if addr = conf2.Csv().IpNew2Old[addr]; addr != "" {
		http2.CallRpc("http://"+addr, enum.Rpc_save_backup_file, func(buf *common.NetPack) {
			buf.WriteString(uid)
			buf.WriteString(pf_id)
		}, func(recvBuf *common.NetPack) {
			if cnt := recvBuf.ReadUInt32(); cnt > 0 {
				files = make([]string, cnt)
				for i := uint32(0); i < cnt; i++ {
					dir := recvBuf.ReadString()
					name := recvBuf.ReadString()
					data := recvBuf.ReadLenBuf()
					files[i] = name
					if fd, e := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); e == nil {
						_, e = fd.Write(data)
						fd.Close()
					}
				}
			}
		})
	}
	if len(files) > 0 {
		addr := fmt.Sprintf("http://%s:%d/save_backup", meta.G_Local.OutIP, meta.G_Local.HttpPort)
		tmp := backupAck{files, pf_id, uid, addr}
		t, _ := template.New("").Parse(kTemplate)
		t.Execute(w, &tmp)
	} else {
		w.Write([]byte("none backup data"))
	}
}
func Http_save_backup(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pf_id := q.Get("pf_id")
	uid := q.Get("uid")
	filename := q.Get("filename")
	ack := "unknow error"
	ptr := &logic.TSaveData{Key: logic.GetSaveKey(pf_id, uid)}
	if ok, _ := dbmgo.Find(logic.KDBSave, "_id", ptr.Key, ptr); !ok {
		ack = "none save data"
	} else if ptr.RollBack(filename) != err.Success {
		ack = "fail to rollback"
	} else {
		ack = "save_rollback: ok"
	}
	w.Write(common.S2B(ack))
}
