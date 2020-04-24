package web

import (
	"common"
	"common/file"
	"common/std"
	"encoding/json"
	"fmt"
	"generate_out/rpc/enum"
	"io/ioutil"
	"nets/http"
	"os"
	"path/filepath"
	"strings"
)

type TCommon struct {
	LocalAddr  string   //GM地址
	CenterList []string //center地址
	SdkAddrs   []string //支付sdk暂时各项目共用
}
type TemplateData struct {
	*TCommon
	GameName string
	Logins   []TLogin //0位空，1起始
	pf_id    []std.StrPair
}
type TLogin struct {
	Name  string
	Addrs []string //登录节点同质的
	Games []TGame
	Gates []std.KeyPair
}
type TGame struct {
	ID        int
	GameAddr  string
	SaveAddrs []string //同节点下的save同质的
}

func (self *TemplateData) GetAddrs() {
	// 拉游戏大区
	http.CallRpc(self.CenterList[0], enum.Rpc_meta_list, func(buf *common.NetPack) {
		buf.WriteString(self.GameName)
	}, func(recvBuf *common.NetPack) {
	LOOP:
		for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
			recvBuf.ReadInt() //svrId
			outip := recvBuf.ReadString()
			port := recvBuf.ReadUInt16()
			svrName := recvBuf.ReadString()
			addr := http.Addr(outip, port)
			for i := 0; i < len(self.Logins); i++ {
				if self.Logins[i].Name == svrName {
					self.Logins[i].Addrs = append(self.Logins[i].Addrs, addr)
					continue LOOP
				}
			}
			self.Logins = append(self.Logins, TLogin{
				Name:  svrName,
				Addrs: []string{addr},
			})
		}
	})
	for i := 0; i < len(self.Logins); i++ {
		if pLogin := &self.Logins[i]; len(pLogin.Addrs) > 0 {
			// 拉大区下所有game
			pLogin.Games = pLogin.Games[:0] //清空旧值
			http.CallRpc(pLogin.Addrs[0], enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("game")
			}, func(recvBuf *common.NetPack) {
				for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
					svrId := recvBuf.ReadInt()
					outip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //name
					pLogin.Games = append(pLogin.Games, TGame{
						ID:       svrId,
						GameAddr: http.Addr(outip, port),
					})
				}
				// 再拉game下所有save
				for i := 0; i < len(pLogin.Games); i++ {
					ptr := &pLogin.Games[i]
					ptr.SaveAddrs = ptr.SaveAddrs[:0] //清空旧值
					http.CallRpc(ptr.GameAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
						buf.WriteString("save")
					}, func(recvBuf *common.NetPack) {
						for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
							recvBuf.ReadInt() //svrId
							outip := recvBuf.ReadString()
							port := recvBuf.ReadUInt16()
							recvBuf.ReadString() //name
							ptr.SaveAddrs = append(ptr.SaveAddrs, http.Addr(outip, port))
						}
					})
				}
			})
			// 拉大区下所有gateway
			pLogin.Gates = pLogin.Gates[:0] //清空旧值
			http.CallRpc(pLogin.Addrs[0], enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("gateway")
			}, func(recvBuf *common.NetPack) {
				for cnt, i := recvBuf.ReadByte(), byte(0); i < cnt; i++ {
					svrId := recvBuf.ReadInt()
					outip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //name
					pLogin.Gates = append(pLogin.Gates, std.KeyPair{http.Addr(outip, port), svrId})
				}
			})
		}
	}
	// 拉到的数据写本地文件
	if fi, e := file.CreateFile("log/", self.GameName+".addr", os.O_TRUNC|os.O_WRONLY); e == nil {
		buf, _ := common.T2B(self)
		fi.Write(buf)
		fi.Close()
	}
	buf, _ := json.MarshalIndent(self, "", "     ")
	fmt.Println(common.B2S(buf))
}
func (self *TemplateData) LoadAddrs() (ret bool) {
	if f, e := os.Open("log/" + self.GameName + ".addr"); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			common.B2T(buf, self)
			buf, _ = json.MarshalIndent(self, "", "     ")
			fmt.Println(common.B2S(buf))
			ret = true
		}
		f.Close()
	}
	return
}

func UpdateHtmls(dirIn, dirOut string, ptr interface{}) { //填充模板，生成可用的HTML文件，方便查看
	if names, err := file.WalkDir(kTemplateDir+dirIn, ".html"); err == nil {
		for _, name := range names {
			fmt.Println("UpdateHtmls:", name)
			out := strings.Replace(name, "template/"+dirIn, dirOut, -1)
			outDir, outName := filepath.Split(out)
			f, _ := file.CreateFile(outDir, outName, os.O_WRONLY|os.O_TRUNC)
			if e := file.TemplateParse(ptr, name, f); e != nil {
				fmt.Println("parse template error: ", e.Error())
			}
			f.Close()
		}
	}
}
func UpdateHtml(fileIn, fileOut string, ptr interface{}) {
	fmt.Println("UpdateHtml:", fileIn)
	in := kTemplateDir + fileIn + ".html"
	out := strings.Replace(in, "template/"+fileIn, fileOut, -1)
	outDir, outName := filepath.Split(out)
	f, _ := file.CreateFile(outDir, outName, os.O_WRONLY|os.O_TRUNC)
	if e := file.TemplateParse(ptr, in, f); e != nil {
		fmt.Println("parse template error: ", e.Error())
	}
	f.Close()
}

// ------------------------------------------------------------
// template func
func (self *TemplateData) AddrLogin(idx int) string {
	if idx < len(self.Logins) {
		return self.Logins[idx].Addrs[0]
	}
	return ""
}
func (self *TemplateData) AddrGame(idx, id int) string {
	if idx < len(self.Logins) {
		for _, v := range self.Logins[idx].Games {
			if v.ID == id {
				return v.GameAddr
			}
		}
	}
	return ""
}
func (self *TemplateData) AddrSave(idx, id int) string {
	if idx < len(self.Logins) {
		for _, v := range self.Logins[idx].Games {
			if v.ID == id && len(v.SaveAddrs) > 0 {
				return v.SaveAddrs[0]
			}
		}
	}
	return ""
}
func (self *TemplateData) PfidSave() (ret []std.StrPair) { //存档互通的渠道
	for _, p := range self.pf_id {
		p.K = strings.ReplaceAll(p.K, " ", "") //剔除空格
		p.K = strings.Split(p.K, ",")[0]
		if p.V != "" {
			p.V = fmt.Sprintf("%s（%s）", p.K, p.V)
		} else {
			p.V = p.K
		}
		ret = append(ret, p)
	}
	return
}
func (self *TemplateData) PfidAll() (ret []std.StrPair) {
	for _, p := range self.pf_id {
		p.K = strings.ReplaceAll(p.K, " ", "") //剔除空格
		if ks := strings.Split(p.K, ","); len(ks) > 1 {
			vs := strings.Split(p.V, ",")
			for i, k := range ks[1:] {
				p.K = k
				p.V = fmt.Sprintf("%s（%s）", p.K, vs[i])
				ret = append(ret, p)
			}
		} else {
			if p.K = ks[0]; p.V != "" {
				p.V = fmt.Sprintf("%s（%s）", p.K, p.V)
			} else {
				p.V = p.K
			}
			ret = append(ret, p)
		}
	}
	return
}
func (self *TemplateData) SplitLogin1() string { //分割国区、海外
	var list []string
	for _, v := range self.Logins {
		if strings.Index(v.Name, "China") >= 0 {
			list = append(list, v.Addrs[0])
		}
	}
	return strings.Join(list, " ")
}
func (self *TemplateData) SplitLogin2() (ret []TLogin) {
	for _, v := range self.Logins {
		if strings.Index(v.Name, "China") < 0 {
			ret = append(ret, v)
		}
	}
	return
}
