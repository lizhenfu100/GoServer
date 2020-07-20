/***********************************************************************
* @ c++ enum 映射到 c# lua
* @ brief
    1、c#替换：enum class -> public enum
	2、lua：
		· 识别出 _Tag 计算出每个枚举值

* @ author zhoumf
* @ date 2020-7-1
***********************************************************************/
package main

import (
	"bytes"
	"common"
	"common/file"
	"common/std"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const (
	kSvrBattleEnum = "../../CXServer/src/svr_battle/Room/Enum/"
	kSvrBattleLua  = "../../CXServer/lua/"
	kClientEnum    = "../../GameClient/Assets/RGScript/Enum/"
	kEnumTemplate  = `local {{.Name}} = {
	{{range .KV}}{{.Name}} = {{.ID}},
	{{end}}
}
`
)

var kFileList = []string{
	"../../CXServer/src/svr_battle/Room/Combat/CalcAttr.h",
	"../../CXServer/src/svr_battle/Room/Buff/Buff.h",
}

// enum class 改为 public enum
func generateBattleEnum() {
	names, _ := file.WalkDir(kSvrBattleEnum, ".h")
	for _, fileName := range names {
		f, _ := os.Open(fileName)
		buf, e := ioutil.ReadAll(f)
		if f.Close(); e == nil {
			s := common.B2S(buf)
			s = strings.ReplaceAll(s, "enum class", "public enum")
			buf = common.S2B(s)
			dir, name := filepath.Split(kClientEnum + strings.TrimPrefix(fileName, kSvrBattleEnum))
			if f, e = file.CreateFile(dir, name, os.O_WRONLY|os.O_TRUNC); e == nil {
				_, e = f.Write(buf)
				if f.Close(); e != nil {
					fmt.Println("Error: ", e.Error())
				}
			}
		}
	}
}

// lua
type LuaEnum struct {
	Lua  string //lua文件路径
	Name string
	KV   []std.KeyPair
	idx  int //枚举值
}

var _enum = regexp.MustCompile(`\w+\s*=?\s*[0-9]*,`)

func (p *LuaEnum) generateLuaEnum() {
	//names, _ := file.WalkDir(kSvrBattleEnum, ".h")
	for _, fileName := range kFileList {
		p.generateLuaEnum1(fileName)
	}
}
func (p *LuaEnum) generateLuaEnum1(fileName string) {
	enumBegin := false
	file.ReadLine(fileName, func(line string) {
		if i := strings.Index(line, "enum "); i >= 0 {
			enumBegin = true
			p.parseClass(line)
		} else if enumBegin {
			if strings.Index(line, "};") >= 0 {
				enumBegin = false
				p.outputLua() //一个枚举完毕，输出到lua
			} else {
				if i := strings.Index(line, "//"); i >= 0 {
					line = line[:i] //剔除注释
				}
				p.parseEnum(line)
			}
		}
	})
}

func (p *LuaEnum) parseClass(s string) {
	for _, e := range [...]string{"enum class ", "enum "} {
		if i := strings.Index(s, e); i >= 0 {
			if j := strings.Index(s[i+len(e):], " "); j >= 0 {
				p.Name = s[i+len(e) : i+len(e)+j]
				if i := strings.Index(s, "//"); i >= 0 {
					p.Lua = s[i+2:] //lua文件路径
				}
				return
			}
		}
	}
}
func (p *LuaEnum) parseEnum(s string) {
	if _enum.MatchString(s) {
		if enum := regexp.MustCompile(`\w+`).FindAllString(s, -1)[0]; enum[0] != '_' {
			if i := strings.Index(s, "="); i >= 0 {
				_idx := s[i+2 : strings.Index(s, ",")]
				if n, e := strconv.Atoi(_idx); e == nil {
					p.KV = append(p.KV, std.KeyPair{enum, n})
					p.idx = n + 1
					return
				}
			}
			p.KV = append(p.KV, std.KeyPair{enum, p.idx})
			p.idx++
		}
	}
}
func (p *LuaEnum) outputLua() {
	//生成lua枚举
	var buf bytes.Buffer
	tpl, err := template.New(p.Lua).Parse(kEnumTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = tpl.Execute(&buf, p)
	//删旧枚举
	f, _ := os.OpenFile(kSvrBattleLua+p.Lua, os.O_RDWR, 0666)
	b, _ := ioutil.ReadAll(f)
	f.Close()
	s := common.B2S(b)
	if i := strings.Index(s, p.Name); i >= 0 {
		if j := strings.Index(s[i:], "}"); j >= 0 {
			s = s[:i-len(`local `)] + s[i+j+2:]
		}
	}
	buf.WriteString(s)
	f, _ = os.OpenFile(kSvrBattleLua+p.Lua, os.O_RDWR|os.O_TRUNC, 0666)
	f.Write(buf.Bytes())
	f.Close()
	//清空缓存，待扫描下份枚举
	p.Lua = ""
	p.Name = ""
	p.KV = p.KV[:0]
	p.idx = 0
}
