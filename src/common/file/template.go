package file

import (
	"bytes"
	"os"
	"text/template"
)

// 内置模板函数：template/funcs.go -> builtins: index/len/...
func CreateTemplate(ptr interface{}, outDir, outName, tempText string) {
	tpl, err := template.New(outName).Parse(tempText)
	if err != nil {
		panic(err.Error())
		return
	}
	f, err := CreateFile(outDir, outName, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		panic(err.Error())
		return
	}
	if err = tpl.Execute(f, ptr); err != nil {
		panic(err.Error())
		return
	}
	f.Close()
}

func ParseTemplate(ptr interface{}, fullname string) ([]byte, error) {
	var bf bytes.Buffer
	if t, e := template.ParseFiles(fullname); e != nil {
		return nil, e
	} else if e = t.Execute(&bf, ptr); e != nil {
		return nil, e
	} else {
		return bf.Bytes(), nil
	}
}
