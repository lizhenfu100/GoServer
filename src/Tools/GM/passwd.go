package main

import (
	"common"
	"common/timer"
	"common/tool/email"
	"fmt"
	"gamelog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	g_passwd string
	g_emails = []string{
		"515693380@qq.com",  //许嘉琪
		"707723219@qq.com",  //杨添怿
		"2370159093@qq.com", //单泽永
		"734688714@qq.com",  //周梦飞
	}
)

func UpdatePasswd() {
	rand.Seed(time.Now().Unix())
	g_passwd = strconv.Itoa(rand.Intn(100000000))
	gamelog.Info("Passwd: " + g_passwd)

	//本周还剩多少时间
	wday := int(time.Now().Weekday()+6) % 7 // weekday but Monday = 0.
	leftSec := wday*24*3600 + timer.TodayLeftSec()
	time.AfterFunc(time.Duration(leftSec)*time.Second, UpdatePasswd)

	for _, v := range g_emails {
		if e := email.SendMail("凉屋GM密码", v, g_passwd, ""); e != nil {
			fmt.Println(e.Error())
		}
	}
}

func Http_check_passwd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	passwd := q.Get("passwd")

	if passwd == g_passwd {
		if f, e := os.Open(kFileDirRoot + "index.html"); e == nil {
			if buf, e := ioutil.ReadAll(f); e == nil {
				w.Write(buf)
			}
			f.Close()
		}
	} else {
		w.Write(common.ToBytes("Passwd error.\nPlease accept the password in your email."))
	}
}
