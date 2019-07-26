package email

import (
	"conf"
	"reflect"
)

var G_EmailCsv = map[string]*csvEmail{}

type csvEmail struct { // Notice：用支持UTF-8的编辑器写csv，否则容易乱码
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
}

func Translate(title, language string) (string, bool) {
	ret, ok := translate(title, language)
	if !ok {
		ret, ok = translate(title, conf.SvrCsv.EmailLanguage)
	}
	return ret, ok
}
func translate(title, language string) (string, bool) {
	if csv, ok := G_EmailCsv[title]; ok {
		ref := reflect.ValueOf(csv).Elem()
		if v := ref.FieldByName(language); v.IsValid() && v.String() != "" {
			return v.String(), true
		}
	}
	return title, false
}
