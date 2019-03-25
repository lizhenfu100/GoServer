package email

import (
	"reflect"
)

var G_Email = make(map[string]*csvEmail)

type csvEmail struct { // Notice：用支持UTF-8的编辑器写csv，否则容易乱码
	Title   string
	En      string
	Zh      string
	Zh_Hant string
	Jp      string
	Ru      string //俄语
	Kr      string //韩语
	Es      string //西班牙语
	Pt_Br   string //葡萄牙语
	Fr      string //法语
	Id      string //印尼语
	De      string //德语
}

func translate(title, language string) string {
	if csv, ok := G_Email[title]; ok {
		ref := reflect.ValueOf(csv).Elem()
		if v := ref.FieldByName(language); v.IsValid() {
			return v.String()
		}
	}
	return ""
}
