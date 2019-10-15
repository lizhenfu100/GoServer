package file

import (
	"os"
	"strings"
	"time"
)

//【会遍历目录，不应频繁调用】
func DelExpired(dir, prefix string, nday /*多少天算过期*/ int) {
	if f, err := os.Open(dir); err == nil {
		list, err := f.Readdir(-1)
		f.Close()
		if err == nil {
			expireTime := time.Now().Add(time.Duration(-nday) * time.Hour * 24)
			for _, fi := range list {
				if strings.HasPrefix(fi.Name(), prefix) && fi.ModTime().Before(expireTime) {
					os.Remove(dir + fi.Name())
				}
			}
		}
	}
}
