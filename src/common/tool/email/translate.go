package email

import (
	"conf"
	"reflect"
)

//go:generate D:\server\bin\gen_conf.exe email emailCsv invalidCsv
type emailCsv map[string]*struct { // Notice：用支持UTF-8的编辑器写csv，否则容易乱码
	Title   string
	En      string
	Zh      string //简中
	Zh_Hant string //繁中
	Jp      string
	Ru      string //俄语
	Kr      string //韩语
	Es      string //西班牙语
	Pt_Br   string //葡萄牙语
	Fr      string //法语
	Id      string //印尼语
	De      string //德语
	Ar      string //阿拉伯语
	Fa      string //波斯语
}
type invalidCsv map[string]*struct {
	Addr string
}

func Translate(title, language string) (string, bool) {
	ret, ok := translate(title, language)
	if !ok {
		ret, ok = translate(title, conf.SvrCsv().EmailLanguage)
	}
	return ret, ok
}
func translate(title, language string) (string, bool) {
	if csv, ok := EmailCsv()[title]; ok {
		ref := reflect.ValueOf(csv).Elem()
		if v := ref.FieldByName(language); v.IsValid() && v.String() != "" {
			return v.String(), true
		}
	}
	return title, false
}

func Invalid(addr string) bool {
	_, ok := InvalidCsv()[addr]
	return ok
}
