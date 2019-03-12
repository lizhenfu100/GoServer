package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	_ "svr_client/test/init"
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
	if resp, err := http.Get(u.String()); err == nil {
		buf, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			fmt.Println(string(buf))
		}
	}
}
