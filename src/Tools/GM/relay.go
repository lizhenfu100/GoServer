package main

import (
	"common"
	"common/file"
	"common/std/sign"
	"generate_out/err"
	"generate_out/rpc/enum"
	"io/ioutil"
	"net/http"
	"net/url"
	mhttp "nets/http"
	"shared_svr/svr_login/gm"
	"strconv"
	"time"
)

func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、追加参数
	k, v := q.Get("k"), q.Get("v")
	pwd := q.Get("pwd")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(k+v+pwd+flag))
	for i := 0; i < len(g_common.CenterList); i++ {
		//2、创建url
		u, _ := url.Parse(g_common.CenterList[i] + "/reset_password")
		//3、生成完整url
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil && i == 0 {
			w.Write(buf)
		}
	}
}
func Http_bind_info_force(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	//1、追加参数
	aid, k, v := q.Get("aid"), q.Get("k"), q.Get("v")
	flag := strconv.FormatInt(time.Now().Unix(), 10)
	q.Set("flag", flag)
	q.Set("sign", sign.CalcSign(aid+k+v+flag))
	for i := 0; i < len(g_common.CenterList); i++ {
		//2、创建url
		u, _ := url.Parse(g_common.CenterList[i] + "/bind_info_force")
		//3、生成完整url
		u.RawQuery = q.Encode()
		if buf := mhttp.Client.Get(u.String()); buf != nil && i == 0 {
			w.Write(buf)
		}
	}
}

func Http_relay_gm_cmd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	addr := q.Get("addr")
	cmd := q.Get("cmd")
	args := q.Get("args")

	mhttp.CallRpc("http://"+addr, enum.Rpc_gm_cmd, func(buf *common.NetPack) {
		buf.WriteString(cmd)
		buf.WriteString(args)
	}, func(recvBuf *common.NetPack) {
		str := recvBuf.ReadString()
		w.Write(common.S2B(str))
	})
}

// ------------------------------------------------------------
// 批量转发
func Http_relay_to_save(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")

	if p, ok := g_map[gameName]; !ok {
		w.Write(common.S2B("GameName error"))
	} else {
		var acks [][]byte
		for _, v := range p.Logins {
			for _, v2 := range v.Games {
				for _, v3 := range v2.SaveAddrs {
					u, _ := url.Parse(v3 + r.RequestURI) //除去域名或ip的url
					if buf := mhttp.Client.Get(u.String()); buf != nil {
						acks = append(acks, buf)
					}
				}
			}
		}
		writeRelayResult(w, acks)
	}
}
func Http_relay_to_login(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gameName := q.Get("game")

	if p, ok := g_map[gameName]; !ok {
		w.Write(common.S2B("GameName error"))
	} else {
		var acks [][]byte
		for _, v := range p.Logins {
			if len(v.Addrs) > 0 {
				u, _ := url.Parse(v.Addrs[0] + r.RequestURI) //除去域名或ip的url
				if buf := mhttp.Client.Get(u.String()); buf != nil {
					acks = append(acks, buf)
				}
			}
		}
		writeRelayResult(w, acks)
	}
}
func writeRelayResult(w http.ResponseWriter, acks [][]byte) {
	//检查回复是否一致
	isSame := true
	for i := 0; i < len(acks); i++ {
		for j := i + 1; j < len(acks); j++ {
			if common.B2S(acks[i]) != common.B2S(acks[j]) {
				isSame = false
			}
		}
	}
	if isSame {
		w.Write(acks[0])
	} else {
		w.Write(common.S2B("Different result !!!\n\n"))
		for i := 0; i < len(acks); i++ {
			w.Write(acks[i])
		}
	}
}

// ------------------------------------------------------------
// 生成礼包码
func Http_gift_code_spawn(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	count, _ := strconv.Atoi(q.Get("count"))

	gm.MakeGiftCodeCsv(key, count)

	r.URL.Path = gm.KGiftCodeDir + key + ".csv"
	svr := http.FileServer(http.Dir("."))
	svr.ServeHTTP(w, r)
}

// ------------------------------------------------------------
// 云存档
func Http_download_save_data(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) //multipart/form-data格式，用于传文件
	saveAddr := "http://" + r.Form.Get("addr")
	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")
	mac := r.Form.Get("mac")
	clientVersion := r.Form.Get("version")
	mhttp.CallRpc(saveAddr, enum.Rpc_save_download_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString(mac)
		buf.WriteString("") //sign
		buf.WriteString(clientVersion)
	}, func(backBuf *common.NetPack) {
		if e := backBuf.ReadUInt16(); e == err.Success {
			w.Write(backBuf.LeftBuf())
		} else {
			w.Write(common.S2B("下载失败Err: " + strconv.Itoa(int(e))))
		}
	})
}
func Http_upload_save_data(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) //multipart/form-data格式，用于传文件
	saveAddr := "http://" + r.Form.Get("addr")
	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")
	mac := r.Form.Get("mac")
	clientVersion := r.Form.Get("version")

	f, _, e := r.FormFile("save")
	if e != nil {
		w.Write(common.S2B("上传失败 无效文件"))
		return
	}
	data, e := ioutil.ReadAll(f)
	f.Close()
	if e != nil {
		w.Write(common.S2B("上传失败 无效文件"))
		return
	}
	mhttp.CallRpc(saveAddr, enum.Rpc_save_upload_binary, func(buf *common.NetPack) {
		buf.WriteString(uid)
		buf.WriteString(pf_id)
		buf.WriteString(mac)
		buf.WriteString("") //sign
		buf.WriteString("") //extra
		buf.WriteLenBuf(data)
		buf.WriteString(clientVersion)
	}, func(backBuf *common.NetPack) {
		errCode := backBuf.ReadUInt16()
		if errCode == err.Success {
			w.Write(backBuf.LeftBuf())
			if e := backBuf.ReadUInt16(); e == err.Success {
				w.Write(common.S2B("ok"))
			} else {
				w.Write(common.S2B("上传失败Err: " + strconv.Itoa(int(e))))
			}
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

	//依次收集center、login、game、save延时信息
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
	if b, e := file.ParseTemplate(rets, kTemplateDir+"ack/net_delay.html"); e == nil {
		w.Write(b)
	} else {
		w.Write(common.S2B(e.Error()))
	}
}
func testDelay(addr string) (ret string) {
	temp := time.Now()
	mhttp.CallRpc(addr, enum.Rpc_meta_list, func(*common.NetPack) {
	}, func(*common.NetPack) {
		ret = time.Now().Sub(temp).String()
	})
	return
}
