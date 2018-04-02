package logic

import (
	"bytes"
	"compress/gzip"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"svr_sdk/msg"
)

const kDBTableName = "Save"

type TSaveData struct {
	Key  string `bson:"_id"` // Pf_id + Uid
	Data string //json
}

// -------------------------------------
// -- Rpc
func Http_download_save_data(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//! 创建回复
	ack := new(msg.Json_ack)
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		{ //数据压缩
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			gw.Write(b)
			gw.Flush()
			gw.Close()
			b = buf.Bytes()
		}
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if r.Form.Get("sign") != msg.CalcSign(s) {
		ack.Retcode = -2
		gamelog.Error("download_save_data: sign failed")
		return
	}

	key := pf_id + "_" + uid
	if ptr, ok := LoadFromDB(key); ok {
		ack.Retcode = 0
		ack.Data = ptr.Data
	}
}

type upload_data struct {
	Uid   string `json:"uid"`
	Pf_id string `json:"pf_id"`
	Data  string `json:"data"`
	Sign  string `json:"sign"`
}

func Http_upload_save_data(w http.ResponseWriter, r *http.Request) {
	buffer := make([]byte, r.ContentLength)
	r.Body.Read(buffer)
	var info upload_data
	{ //数据解压
		if gr, err := gzip.NewReader(bytes.NewReader(buffer)); err == nil {
			if buf, err := ioutil.ReadAll(gr); err == nil {
				json.Unmarshal(buf, &info)
			}
			gr.Close()
		}
	}

	//! 创建回复
	ack := new(msg.Retcode_ack)
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(b)
		gamelog.Debug("ack: %v", ack)
	}()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s&data=%s", info.Uid, info.Pf_id, info.Data)
	if info.Sign != msg.CalcSign(s) {
		ack.Retcode = -2
		gamelog.Error("upload_save_data: sign failed")
		return
	}

	key := info.Pf_id + "_" + info.Uid
	ptr, ok := LoadFromDB(key)
	ptr.Key = key
	ptr.Data = info.Data
	if ok {
		dbmgo.UpdateToDB(kDBTableName, bson.M{"_id": ptr.Key}, bson.M{"$set": bson.M{"data": ptr.Data}})
	} else {
		dbmgo.InsertToDB(kDBTableName, ptr)
	}
	ack.Retcode = 0
}

// -------------------------------------
// -- 辅助函数
func LoadFromDB(key string) (*TSaveData, bool) {
	data := new(TSaveData)
	ok := dbmgo.Find(kDBTableName, "_id", key, data)
	return data, ok
}
