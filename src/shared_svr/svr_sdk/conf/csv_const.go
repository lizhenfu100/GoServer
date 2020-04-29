package conf

//go:generate D:\server\bin\gen_conf.exe pingxxSub conf
type pingxxSub map[string]*struct {
	GamePf string
	AppId  string
}
