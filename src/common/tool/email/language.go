package email

// Notice：用支持UTF-8的编辑器操作，否则容易乱码
var G_Email = make(map[string]*csvEmail)

type csvEmail struct {
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
