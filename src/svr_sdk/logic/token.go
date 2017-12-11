package logic

func CheckToken(channel string) bool {
	if info, ok := g_token_map[channel]; ok {
		return info.checkFunc()
	}
	return false
}

type tokenInfo struct {
	token     string
	checkFunc func() bool
}

var g_token_map = map[string]*tokenInfo{
	"360":  {"aaasdfe", _check_360},
	"mini": {"bbbaads", _check_mini},
}

func _check_360() bool {
	return true
}
func _check_mini() bool {
	return true
}
