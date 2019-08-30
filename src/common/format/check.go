package format

import (
	"regexp"
)

//32长，任意非空字符
func CheckPasswd(s string) bool {
	ok, _ := regexp.MatchString(`^\S{3,32}$`, s)
	return ok
}

func CheckBindValue(key, s string) (ok bool) { //格式须不一样，防止用户混用
	switch key {
	case "name": //非纯数字、32长：数字、字母、下划线、横杠、点、*
		if ok, _ = regexp.MatchString(`^[0-9]*$`, s); ok {
			return false
		}
		ok, _ = regexp.MatchString(`^[\w-\.\*]{3,32}$`, s)
	case "email":
		ok, _ = regexp.MatchString(`^[\w-\.]+@[\w-]+(\.[\w-]+)+$`, s)
	case "phone":
		ok, _ = regexp.MatchString(`^[0-9]{6,14}$`, s)
	default:
		ok = false
	}
	return
}
