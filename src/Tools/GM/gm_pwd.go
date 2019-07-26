package main

import (
	"common"
	"common/timer"
	"gamelog"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var g_passwd string

func UpdatePasswd() {
	g_passwd = strconv.Itoa(rand.Intn(100000000))
	gamelog.Info("Passwd: " + g_passwd)

	//本周还剩多少时间
	wday := int(time.Now().Weekday()+6) % 7 // weekday but Monday = 0.
	leftSec := wday*24*3600 + timer.TodayLeftSec()
	time.AfterFunc(time.Duration(leftSec)*time.Second, UpdatePasswd)
}

func Http_check_passwd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("passwd") == g_passwd {
		if f, e := os.Open(kFileDirRoot + "index.html"); e == nil {
			if buf, e := ioutil.ReadAll(f); e == nil {
				w.Write(buf)
			}
			f.Close()
		}
	} else {
		w.Write(common.S2B("Passwd error.\nPlease accept the password in your email."))
	}
}
