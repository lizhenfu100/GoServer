package web

import (
	"bytes"
	"common"
	"common/file"
	"common/std/sign"
	"generate_out/err"
	"generate_out/rpc/enum"
	"io/ioutil"
	"net/http"
	"net/url"
	mhttp "nets/http"
	"shared_svr/svr_login/gift_bag"
	"strconv"
	"strings"
	"time"
)

func GetAddrs(q url.Values) (ret []string) {
	list := q["_list"]
	q.Del("_list")
	for i := 0; i < len(list); i++ {
		for _, v := range strings.Split(list[i], " ") {
			if v != "" {
				ret = append(ret, v)
			}
		}
	}
	return
}
func foreach_svr(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	for _, v := range GetAddrs(q) {
		u, _ := url.Parse(v + r.URL.Path)
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil {
			w.Write(buf)
			w.Write([]byte("\n"))
		} else {
			w.Write([]byte("http fail\n"))
		}
	}
}
func mergeAcks(addrs []string, rawQuery string, path string, w http.ResponseWriter) {
	var acks [][]byte
	for _, v := range addrs {
		u, _ := url.Parse(v + path)
		u.RawQuery = rawQuery
		if buf := mhttp.Client.Get(u.String()); buf != nil {
			acks = append(acks, buf)
		} else {
			acks = append(acks, []byte("http fail\n"))
		}
	}
	isSame := true //检查回复是否一致
	for i := 0; i < len(acks); i++ {
		for j := i + 1; j < len(acks); j++ {
			if common.B2S(acks[i]) != common.B2S(acks[j]) {
				isSame = false
			}
		}
	}
	if len(acks) == 0 {
		w.Write([]byte("尚未选取大区或节点"))
	} else if isSame {
		w.Write(acks[0])
	} else {
		w.Write([]byte("Error: Different result !!!\n\n"))
		for i := 0; i < len(acks); i++ {
			w.Write(acks[i])
		}
	}
}
func relay_to(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addrs := GetAddrs(q)
	mergeAcks(addrs, q.Encode(), r.URL.Path, w)
}
func relay_to_save(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if p, ok := g_map[q.Get("game")]; !ok {
		w.Write([]byte("GameName error"))
	} else {
		var addrs []string
		for _, v := range p.Logins {
			for _, v2 := range v.Games {
				for _, v3 := range v2.SaveAddrs {
					addrs = append(addrs, v3)
				}
			}
		}
		mergeAcks(addrs, r.URL.RawQuery, r.URL.Path, w)
	}
}
func relay_gm_cmd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("addr")
	cmd := q.Get("cmd")
	args := q.Get("args")
	mhttp.CallRpc("http://"+addr, enum.Rpc_gm_cmd, func(buf *common.NetPack) {
		buf.WriteString(cmd)
		buf.WriteString(args)
	}, func(recvBuf *common.NetPack) {
		w.Write(common.S2B(recvBuf.ReadString()))
	})
}

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	k, v := q.Get("k"), q.Get("v")
	pwd := q.Get("pwd")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(k+v+pwd+flag))
	mergeAcks(g_common.CenterList, q.Encode(), r.URL.Path, w)
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	aid, k, v := q.Get("aid"), q.Get("k"), q.Get("v")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(aid+k+v+flag))
	for _, v := range g_common.CenterList {
		u, _ := url.Parse(v + r.URL.Path)
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil {
			w.Write(buf)
			w.Write([]byte("\n\n"))
		} else {
			w.Write([]byte("http fail\n\n"))
		}
	}
}

// ------------------------------------------------------------
// 生成礼包码
func Http_gift_code_spawn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	game := q.Get("game")
	count, _ := strconv.Atoi(q.Get("count"))

	gift_bag.MakeGiftCodeCsv(game, key, count)

	if q.Get("all") == "" {
		r.URL.Path = gift_bag.KGiftCodeDir + game + "/" + gift_bag.KGiftCodeTemp //只给新礼包码
	} else {
		r.URL.Path = gift_bag.KGiftCodeDir + game + "/" + key + ".csv" //历史所有礼包码
	}
	if count < 1000 {
		var buf bytes.Buffer
		vs, _ := file.ReadCsv(r.URL.Path)
		for _, v := range vs {
			buf.WriteString(v[0])
			buf.WriteString("\n")
		}
		w.Write(buf.Bytes())
	} else {
		http.FileServer(http.Dir(".")).ServeHTTP(w, r)
	}
}

// ------------------------------------------------------------
// 云存档
func Http_download_save_data(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	addr := "http://" + r.Form.Get("addr")
	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")
	mhttp.CallRpc(addr, enum.Rpc_save_gm_dn, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
	}, func(backBuf *common.NetPack) {
		if e := backBuf.ReadUInt16(); e == err.Success {
			w.Write(backBuf.LeftBuf())
		} else {
			w.Write(common.S2B("下载失败Err: " + strconv.Itoa(int(e))))
		}
	})
}
func Http_upload_save_data(w http.ResponseWriter, r *http.Request) {
	f, _, e := r.FormFile("save")
	addr := "http://" + r.Form.Get("addr")
	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")
	if e != nil {
		w.Write([]byte("上传失败 无效文件"))
		return
	}
	data, e := ioutil.ReadAll(f)
	if f.Close(); e != nil {
		w.Write([]byte("上传失败 无效文件"))
		return
	}
	mhttp.CallRpc(addr, enum.Rpc_save_gm_up, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString("") //extra
		buf.WriteLenBuf(data)
	}, func(backBuf *common.NetPack) {
		if e := backBuf.ReadUInt16(); e == err.Success {
			w.Write([]byte("ok"))
		} else {
			w.Write(common.S2B("上传失败Err: " + strconv.Itoa(int(e))))
		}
	})
}

// ------------------------------------------------------------
// 全节点延时
func Http_view_net_delay(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	game := q.Get("game")
	idx_ := q.Get("idx") //大区
	idx, _ := strconv.Atoi(idx_)
	data := g_map[game]

	type RetInfo struct {
		Name string
		Addr string
		Info string
	}
	rets := make([]RetInfo, 0, 16)

	//依次收集center、login、gateway、game、save延时信息
	for k, v := range data.CenterList {
		rets = append(rets, RetInfo{
			"Center" + strconv.Itoa(k),
			v,
			testDelay(v)})
	}
	login := &data.Logins[idx]
	for k, v := range login.Addrs {
		rets = append(rets, RetInfo{
			login.Name + strconv.Itoa(k),
			v,
			testDelay(v)})
	}
	for _, v := range login.Gates {
		rets = append(rets, RetInfo{
			"Gateway" + strconv.Itoa(v.ID),
			v.Name,
			testDelay(v.Name)})
	}
	for _, v := range login.Games {
		rets = append(rets, RetInfo{
			"Game" + strconv.Itoa(v.ID),
			v.GameAddr,
			testDelay(v.GameAddr)})
		for k, v := range v.SaveAddrs {
			rets = append(rets, RetInfo{
				"Save" + strconv.Itoa(k),
				v,
				testDelay(v)})
		}
	}
	//格式化html
	if e := file.TemplateParse(rets, kTemplateDir+"ack/net_delay.html", w); e != nil {
		w.Write(common.S2B(e.Error()))
	}
}
func testDelay(addr string) (ret string) {
	temp := time.Now()
	mhttp.CallRpc(addr, enum.Rpc_timestamp, func(*common.NetPack) {
	}, func(*common.NetPack) {
		ret = time.Now().Sub(temp).String()
	})
	return
}

// ------------------------------------------------------------
// 批量封号
func Http_batch_forbid(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	game := q.Get("game")
	data := g_map[game].Logins
	type RetInfo struct {
		Name string
		Addr string
	}
	rets := make([]RetInfo, 0, 16)
	for i := 0; i < len(data); i++ {
		for _, v := range data[i].Games {
			if v.ID < common.KIdMod { //排除分流节点
				rets = append(rets, RetInfo{
					data[i].Name + strconv.Itoa(v.ID),
					v.GameAddr})
			}
		}
	}
	if e := file.TemplateParse(rets, kTemplateDir+"ack/game_meta.html", w); e != nil {
		w.Write(common.S2B(e.Error()))
	}
}
