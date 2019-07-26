package main

import (
	"common"
	"common/file"
	"encoding/json"
	"fmt"
	"generate_out/rpc/enum"
	"html/template"
	"io/ioutil"
	"nets/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	g_common = TCommon{
		CenterAddr: "http://3.17.67.102:7000",
		CenterList: []string{
			"http://3.16.163.125:7000",
			"http://18.221.148.84:7000",
			"http://18.223.109.103:7000",
			"http://18.216.113.27:7000",
		},
		SdkAddrs: []string{"",
			"http://120.78.152.152:7002", //1 北美
			"http://120.78.152.152:7002", //2 亚洲
			"http://120.78.152.152:7002", //3 欧洲
			"http://120.78.152.152:7002", //4 南美
			"http://120.78.152.152:7002", //5 中国华北
			"http://120.78.152.152:7002", //6 中国华南
		},
	}
	g_map = map[string]TemplateData{
		"HappyDiner": {
			GameName: "HappyDiner", //游戏名，以及对应的大区
			Logins: []TLogin{{Login: ""}, //0位空，大区编号从1起始
				{Login: "http://3.17.67.102:7030"},    //1 北美
				{Login: "http://13.229.215.168:7030"}, //2 亚洲
				{Login: "http://18.185.80.202:7030"},  //3 欧洲
				{Login: "http://54.94.211.178:7030"},  //4 南美
				{Login: "http://39.96.196.250:7030"},  //5 中国华北
				{Login: "http://47.106.35.74:7030"},   //6 中国华南
			},
		},
		"SoulKnight": {
			GameName: "SoulKnight",
			Logins: []TLogin{{Login: ""},
				{Login: ""},                          //1 北美
				{Login: ""},                          //2 亚洲
				{Login: ""},                          //3 欧洲
				{Login: ""},                          //4 南美
				{Login: "http://39.97.111.110:7030"}, //5 中国华北
				{Login: "http://39.108.87.225:7030"}, //6 中国华南
			},
		},
	}
)

type TCommon struct {
	LocalAddr  string
	CenterAddr string
	CenterList []string
	SdkAddrs   []string //支付sdk暂时各项目共用
}
type TLogin struct {
	Login string
	Games map[int]TGame
}
type TGame struct {
	Game  string
	Saves []string
}
type TemplateData struct {
	TCommon
	GameName string
	Logins   []TLogin //0位空，1起始
}

func (self *TemplateData) GetAddrs() {
	// 先拉login下所有game
	for i := 1; i < len(self.Logins); i++ {
		if p1 := &self.Logins[i]; p1.Login != "" {
			p1.Games = map[int]TGame{} //清空旧值
			http.CallRpc(p1.Login, enum.Rpc_meta_list, func(buf *common.NetPack) {
				buf.WriteString("game")
				buf.WriteString("")
			}, func(recvBuf *common.NetPack) {
				cnt := recvBuf.ReadByte()
				for i := byte(0); i < cnt; i++ {
					svrId := recvBuf.ReadInt() //svrId
					outip := recvBuf.ReadString()
					port := recvBuf.ReadUInt16()
					recvBuf.ReadString() //name
					p1.Games[svrId] = TGame{
						Game: http.Addr(outip, port),
					}
				}
				// 再拉game下所有save
				for k, v := range p1.Games {
					tmpk, tmpv := k, v          //rang k v 均是固定地址，不可直接闭包(go是引用捕获)
					tmpv.Saves = tmpv.Saves[:0] //清空旧值
					http.CallRpc(v.Game, enum.Rpc_meta_list, func(buf *common.NetPack) {
						buf.WriteString("save")
						buf.WriteString("")
					}, func(recvBuf *common.NetPack) {
						cnt := recvBuf.ReadByte()
						for i := byte(0); i < cnt; i++ {
							recvBuf.ReadInt() //svrId
							outip := recvBuf.ReadString()
							port := recvBuf.ReadUInt16()
							recvBuf.ReadString() //name
							tmpv.Saves = append(tmpv.Saves, http.Addr(outip, port))
						}
						p1.Games[tmpk] = tmpv
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
func (self *TemplateData) LoadAddrs() bool {
	if f, e := os.Open("log/" + self.GameName + ".addr"); e == nil {
		if buf, e := ioutil.ReadAll(f); e == nil {
			f.Close()
			common.B2T(buf, self)
			buf, _ = json.MarshalIndent(self, "", "     ")
			fmt.Println(common.B2S(buf))
			return true
		}
	}
	return false
}

func UpdateHtmls(dirIn, dirOut string, ptr interface{}) { //填充模板，生成可用的HTML文件，方便查看
	if names, err := file.WalkDir(kFileDirRoot+"template/"+dirIn, ".html"); err == nil {
		for _, name := range names {
			if t, e := template.ParseFiles(name); e != nil {
				fmt.Println("parse template error: ", e.Error())
			} else {
				fmt.Println("UpdateHtmls:", name)
				fullname := strings.Replace(name, "template/"+dirIn, dirOut, -1)
				dir, name := filepath.Split(fullname)
				f, _ := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC)
				if e := t.Execute(f, ptr); e != nil {
					fmt.Println(e.Error())
				}
				f.Close()
			}
		}
	}
}
func UpdateHtml(fileIn, fileOut string, ptr interface{}) {
	fullname := kFileDirRoot + "template/" + fileIn + ".html"
	if t, e := template.ParseFiles(fullname); e != nil {
		fmt.Println("parse template error: ", e.Error())
	} else {
		fmt.Println("UpdateHtmls:", fileIn)
		fullname = strings.Replace(fullname, "template/"+fileIn, fileOut, -1)
		dir, name := filepath.Split(fullname)
		f, _ := file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC)
		if e := t.Execute(f, ptr); e != nil {
			fmt.Println(e.Error())
		}
		f.Close()
	}
}

// ------------------------------------------------------------
// template func
func (self *TemplateData) AddrLogin(id int) string {
	return self.Logins[id].Login
}
func (self *TemplateData) AddrGame(id1, id2 int) string {
	return self.Logins[id1].Games[id2].Game
}
func (self *TemplateData) AddrSave(id1, id2 int) string {
	if v, ok := self.Logins[id1].Games[id2]; ok {
		return v.Saves[0]
	}
	return ""
}
