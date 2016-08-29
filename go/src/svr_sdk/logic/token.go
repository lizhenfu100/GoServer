/***********************************************************************
* @ 约定的Token
* @ brief
    1、各个渠道的token不同，验证方式也可能不一样，需可水平扩展

    2、验证通过后，转给"recv_sdk_msg.go"做业务处理

* @ author zhoumf
* @ date 2016-8-16
***********************************************************************/
package logic

type TokenInfo struct {
	token     string
	checkFunc func() bool
}

var g_token_map = map[string]*TokenInfo{
	"360":  {"aaasdfe", _check_360},
	"mini": {"bbbaads", _check_mini},
}

func CheckToken(channel string) bool {
	if info, ok := g_token_map[channel]; ok {
		return info.checkFunc()
	}
	return false
}

func _check_360() bool {
	return true
}
func _check_mini() bool {
	return true
}
