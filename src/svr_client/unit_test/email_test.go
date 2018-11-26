package unit_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func Test_email(t *testing.T) {
	addr := "http://52.14.1.205:7030/ask_reset_password"
	//addr := "http://127.0.0.1:7030/ask_reset_password"
	name := "2370159093@qq.com"
	//name := "1@qq.com"
	passwd := "123123"

	//1、创建url
	u, _ := url.Parse(addr)
	q := u.Query()
	//2、写入参数
	q.Set("name", name)
	q.Set("passwd", passwd)
	//3、生成完整url
	u.RawQuery = q.Encode()
	if res, err := http.Get(u.String()); err == nil {
		defer res.Body.Close()
		if buf, err := ioutil.ReadAll(res.Body); err == nil {
			fmt.Println(string(buf))
		}
	}
}
