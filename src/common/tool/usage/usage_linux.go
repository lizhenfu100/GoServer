package usage

import (
	"common/tool/wechat"
	"syscall"
)

func Check() {
	if memLess() {
		wechat.SendMsg("内存不足")
	} else if disk() > 90 {
		wechat.SendMsg("磁盘占用过高")
	}
}
func disk() byte {
	var v syscall.Statfs_t
	if syscall.Statfs("/home", &v) == nil {
		return 100 - byte(100*(float64(v.Bfree)/float64(v.Blocks)))
	}
	return 0
}
func memLess() bool {
	var v syscall.Sysinfo_t
	if syscall.Sysinfo(&v) == nil {
		return (v.Freeram/1024)/1024 < 50
	}
	return false
}
