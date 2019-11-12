package web

import (
	"common"
	"common/file"
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
type TLogin struct {
	Name  string
	Addrs []string //登录节点同质的
	Games []TGame
}
type TGame struct {
	ID        int
	GameAddr  string
	SaveAddrs []string //同节点下的save同质的
}
type TemplateData struct {
	*TCommon
	GameName string
	Logins   []TLogin //0位空，1起始
}

func (self *TemplateData) GetAddrs() {
	// 先拉login下所有game
	for i := 1; i < len(self.Logins); i++ {
		if p1 := &self.Logins[i]; len(p1.Addrs) > 0 {
			p1.Games = p1.Games[:0] //清空旧值
			http.CallRpc(p1.Addrs[0], enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("game")
				buf.WriteString("")
			}, func(recvBuf *common.NetPack) {
				cnt := recvBuf.ReadByte()
				for i := byte(0); i < cnt; i++ {
					svrId := recvBuf.ReadInt() //svrId
					outip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //name
					p1.Games = append(p1.Games, TGame{
						ID:       svrId,
						GameAddr: http.Addr(outip, port),
					})
				}
				// 再拉game下所有save
				for i := 0; i < len(p1.Games); i++ {
					ptr := &p1.Games[i]
					ptr.SaveAddrs = ptr.SaveAddrs[:0] //清空旧值
					http.CallRpc(ptr.GameAddr, enum.Rpc_meta_list, func(buf *common.NetPack) {
						buf.WriteString("save")
						buf.WriteString("")
					}, func(recvBuf *common.NetPack) {
						cnt := recvBuf.ReadByte()
						for i := byte(0); i < cnt; i++ {
							recvBuf.ReadInt() //svrId
							outip := recvBuf.ReadString()
							port := recvBuf.ReadUInt16()
							recvBuf.ReadString() //name
							ptr.SaveAddrs = append(ptr.SaveAddrs, http.Addr(outip, port))
						}
					})
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
func (self *TemplateData) AddrLogin(id int) string {
	if v := self.Logins[id].Addrs; len(v) > 0 {
		return v[0]
	}
	return ""
}
func (self *TemplateData) AddrGame(id1, id2 int) string {
	for _, v := range self.Logins[id1].Games {
		if v.ID == id2 {
			return v.GameAddr
		}
	}
	return ""
}
func (self *TemplateData) AddrSave(id1, id2 int) string {
	for _, v := range self.Logins[id1].Games {
		if v.ID == id2 {
			return v.SaveAddrs[0]
		}
	}
	return ""
}
