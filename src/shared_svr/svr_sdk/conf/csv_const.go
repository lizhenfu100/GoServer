package conf

//go:generate D:\server\bin\gen_conf.exe conf pingxxSub thinkingApp
type pingxxSub map[string]*struct {
	GamePf string
	AppId  string
}
type thinkingApp map[string]*struct {
	Game  string
	AppId string
}
