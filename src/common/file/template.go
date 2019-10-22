package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// 内置模板函数：template/funcs.go -> builtins: index/len/...
func CreateTemplate(ptr interface{}, outDir, outName, tempText string) error {
	tpl, err := template.New(outName).Parse(tempText)
	if err != nil {
		return err
	}
	f, err := CreateFile(outDir, outName, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return err
	}
	err = tpl.Execute(f, ptr)
	f.Close()
	return err
}

func TemplateDir(ptr interface{}, dirIn, dirOut, suffix string) { //目录下的模板，输出到另一目录
	if names, e := WalkDir(dirIn, suffix); e == nil {
		for _, name := range names {
			out := strings.Replace(name, dirIn, dirOut, -1)
			outDir, outName := filepath.Split(out)
			f, _ := CreateFile(outDir, outName, os.O_WRONLY|os.O_TRUNC)
			if e = TemplateParse(ptr, name, f); e != nil {
				fmt.Println("parse template error: ", e.Error())
			}
			f.Close()
		}
	}
}

func TemplateParse(ptr interface{}, fileIn string, w io.Writer) error {
	if t, e := template.ParseFiles(fileIn); e != nil {
		return e
	} else {
		return t.Execute(w, ptr)
	}
}
