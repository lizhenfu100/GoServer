/***********************************************************************
* @ 云存档
* @ brief
	1、玩家UID、机器码Mac、存档Data，三者一一对应，上传下载须查重

	2、机器码放本地存档(加密)中，仅建档时调API取
		· 之后网络交互，从存档取

	3、客户端发现【存档机器码、API返回的机器码】不一致，将存档禁用

	4、限制换设备后的下载(每月3次)：机器码变更时

* @ 单机防作弊
	1、后台不断变更密钥，用于金币、钻石、攻击力...敏感数据，防止用户窜改

* @ author zhoumf
* @ date 2018-10-31
***********************************************************************/
package logic

import (
	"bytes"
	"common"
	"common/compress"
	"common/sign"
	"compress/gzip"
	"dbmgo"
	"encoding/json"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"svr_sdk/msg"
	"time"
)

const (
	kDBTable         = "Save"
	kDownCntPerMonth = 1
)

type TSaveData struct {
	Key       string `bson:"_id"` // Pf_id + Uid
	Mac       string //机器码，取自存档，中途不用API取
	Data      []byte
	UpTime    int64
	DownTime  int64
	DownCount byte //限制设备变更的下载(每月n次)
}

func Rpc_save_get_meta_info(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()

	key := pf_id + "_" + uid
	ptr := &TSaveData{}
	if dbmgo.Find(kDBTable, "_id", key, ptr) {
		ack.WriteInt64(ptr.UpTime)
	} else {
		ack.WriteInt64(0)
	}
}
func upload(pf_id, uid, mac string, data []byte) uint16 {
	ptr, key := &TSaveData{}, pf_id+"_"+uid
	if dbmgo.Find(kDBTable, "mac", mac, ptr) {
		if ptr.Key != key {
			return err.Record_repeat
		} else {
			ptr.Data = data
			ptr.UpTime = time.Now().Unix()
			dbmgo.UpdateId(kDBTable, ptr.Key, ptr)
		}
	} else {
		ok := dbmgo.Find(kDBTable, "_id", key, ptr)
		ptr.Key = key
		ptr.Mac = mac
		ptr.Data = data
		ptr.UpTime = time.Now().Unix()
		if ok {
			dbmgo.UpdateId(kDBTable, ptr.Key, ptr)
		} else {
			dbmgo.Insert(kDBTable, ptr)
		}
	}
	//fmt.Println("---------------upload: ", len(ptr.Data), ptr)
	return err.Success
}
func download(pf_id, uid, mac string) (*TSaveData, uint16) {
	ptr, key := new(TSaveData), pf_id+"_"+uid
	if !dbmgo.Find(kDBTable, "_id", key, ptr) {
		return nil, err.Record_cannot_find
	} else if pf_id != "IOS" && ptr.Mac != mac {
		timenow := time.Now().Unix()
		if ptr.DownCount += 1; ptr.DownCount <= kDownCntPerMonth {
			dbmgo.UpdateId(kDBTable, key, bson.M{"$set": bson.M{
				"mac":       mac,
				"downcount": ptr.DownCount}})
		} else if timenow-ptr.DownTime > 3600*24*30 {
			ptr.DownTime = timenow
			ptr.DownCount = 1
			dbmgo.UpdateId(kDBTable, key, bson.M{"$set": bson.M{
				"mac":       mac,
				"downtime":  ptr.DownTime,
				"downcount": ptr.DownCount}})
		} else {
			return nil, err.Record_download_limit
		}
	}
	//fmt.Println("----------------download: ", len(ptr.Data), ptr)
	return ptr, err.Success
}

// -------------------------------------
// -- Json 存档
type upload_data struct {
	Uid   string `json:"uid"`
	Pf_id string `json:"pf_id"` //平台id
	Mac   string `json:"mac"`
	Data  string `json:"data"`
	Sign  string `json:"sign"`
}

func Http_upload_save_data(w http.ResponseWriter, r *http.Request) {
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	if _, e := buf.ReadFrom(r.Body); e != nil {
		gamelog.Error("ReadBody: %s", e.Error())
		return
	}
	var info upload_data
	{ //数据解压
		if gr, e := gzip.NewReader(buf); e == nil {
			if buf, e := ioutil.ReadAll(gr); e == nil {
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
	if info.Sign != sign.CalcSign(s) {
		ack.Retcode = -2
		gamelog.Error("upload_save_data: sign failed")
		return
	}

	errcode := upload(info.Pf_id, info.Uid, info.Mac, []byte(info.Data))
	if errcode == err.Success {
		ack.Retcode = 0
	} else {
		ack.Retcode = int(errcode)
	}
}
func Http_download_save_data(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//! 创建回复
	ack := new(msg.Json_ack)
	ack.Retcode = -1
	defer func() {
		b, _ := json.Marshal(&ack)
		w.Write(compress.Compress(b))
		gamelog.Debug("ack: %v", ack)
	}()

	uid := r.Form.Get("uid")
	pf_id := r.Form.Get("pf_id")
	mac := r.Form.Get("mac")

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if r.Form.Get("sign") != sign.CalcSign(s) {
		ack.Retcode = -2
		gamelog.Error("download_save_data: sign failed")
		return
	}

	if ptr, errcode := download(pf_id, uid, mac); errcode == err.Success {
		ack.Retcode = 0
		ack.Data = string(ptr.Data)
	} else {
		ack.Retcode = int(errcode)
	}
}
func Rpc_save_upload_data(req, ack *common.NetPack) {
}
func Rpc_save_download_data(req, ack *common.NetPack) {
}

// -------------------------------------
// -- Binary 存档
func Rpc_save_upload_binary(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()
	data := req.LeftBuf()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		ack.WriteUInt16(err.Sign_failed)
		gamelog.Error("Rpc_save_upload_binary: sign failed")
		return
	}

	errcode := upload(pf_id, uid, mac, data)
	ack.WriteUInt16(errcode)
}
func Rpc_save_download_binary(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		ack.WriteUInt16(err.Sign_failed)
		gamelog.Error("Rpc_save_download_binary: sign failed")
		return
	}

	if ptr, errcode := download(pf_id, uid, mac); errcode == err.Success {
		ack.WriteUInt16(err.Success)
		ack.WriteBuf(ptr.Data)
	} else {
		ack.WriteUInt16(errcode)
	}
}
