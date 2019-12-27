package format

import (
	"regexp"
)

var (
	_pwd   = regexp.MustCompile(`^\S{3,32}$`)
	_num   = regexp.MustCompile(`^[0-9]*$`)
	_name  = regexp.MustCompile(`^[\w-\.\*]{3,32}$`)
	_email = regexp.MustCompile(`^[\w-\.]+@[\w-]+(\.[\w-]+)+$`)
	_phone = regexp.MustCompile(`^[0-9]{6,14}$`)
)

//32长，任意非空字符
func CheckPasswd(v string) bool { return _pwd.MatchString(v) }

//格式须不一样，防止用户混用
func CheckBindValue(key, v string) bool {
	switch key {
	case "name": //非纯数字、32长：数字、字母、下划线、横杠、点、*
		if _num.MatchString(v) {
			return false
		}
		return _name.MatchString(v)
	case "email":
		return _email.MatchString(v)
	case "phone":
		return _phone.MatchString(v)
	}
	return false
}
