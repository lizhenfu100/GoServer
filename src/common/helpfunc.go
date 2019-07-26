package common

import (
	"strings"
)

func IsMatchVersion(a, b string) bool {
	if a == "" || b == "" {
		return true
	}
	// 空版本号能与任意版本匹配
	// 版本号格式：1.12.233，前两组一致的版本间可匹配，第三组用于小调整、bug修复
	idx := strings.LastIndex(a, ".")
	return a[:idx] == b[:idx]
}

func SwapBuf(a, b *[]byte) { *a, *b = *b, *a }
func ClearBuf(p *[]byte)   { *p = (*p)[:0] } //len(0), cap(old), 旧数据不会修改
