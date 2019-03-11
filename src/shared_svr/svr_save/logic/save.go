/***********************************************************************
* @ 云存档
* @ brief
	1、玩家UID、机器码Mac ：一个账号绑定多个设备，一个设备只能绑一个账号
		· 上传下载须查重，该mac是否自己的

	2、机器码放本地存档(加密)中，仅建档时调API取
		· 之后网络交互，从存档取

	3、客户端发现【存档机器码、API返回的机器码】不一致，将存档禁用

	4、限制设备更换频率
		· 前三次绑定，可任意时间
		· 后续的绑定，一周一次

* @ 单机防作弊
	1、后台不断变更密钥，用于金币、钻石、攻击力...敏感数据，防止用户窜改

* @ author zhoumf
* @ date 2018-10-31
***********************************************************************/
package logic

import (
	"common"
	"common/sign"
	"conf"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

const (
	kDBSave       = "Save"
	kDBMac        = "SaveMac"
	kChangePeriod = 3600 * 24 * 7
)

type TSaveData struct {
	Key    string `bson:"_id"` // Pf_id + Uid
	Data   []byte
	UpTime int64
	ChTime int64  //换设备的时刻
	MacCnt byte   //该玩家绑定的设备数目
	Extra  string //json
}
type MacInfo struct {
	Mac string `bson:"_id"` //机器码，取自存档，中途不用API取
	Key string //Pf_id + Uid
}

func Rpc_save_get_meta_info(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()

	key, ptr := pf_id+"_"+uid, &TSaveData{}
	if ok, e := dbmgo.Find(kDBSave, "_id", key, ptr); ok {
		ack.WriteInt64(ptr.UpTime)
	} else if e != nil {
		ack.WriteInt64(-1)
	} else {
		ack.WriteInt64(0)
	}
}
func Rpc_save_check_mac(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()

	errCode := checkMac(pf_id, uid, mac)
	ack.WriteUInt16(errCode)
}
func checkMac(pf_id, uid, mac string) uint16 { //Notice：不可调换错误码判断顺序
	pSave, pMac := &TSaveData{Key: pf_id + "_" + uid}, &MacInfo{}
	okMac, _ := dbmgo.Find(kDBMac, "_id", mac, pMac)
	if okMac {
		if pMac.Key != pSave.Key {
			gamelog.Info("Record_mac_already_bind: mac(%s) new(%s) old(%s)", mac, pSave.Key, pMac.Key)
			return err.Record_mac_already_bind
		}
	}
	if okSave, _ := dbmgo.Find(kDBSave, "_id", pSave.Key, pSave); !okSave {
		return err.Record_cannot_find
	} else if !okMac && pSave.MacCnt >= conf.SvrCsv.FreeBindMacMax { //新设备绑定受限
		if now := time.Now().Unix(); now-pSave.ChTime < kChangePeriod {
			return err.Record_bind_limit
		}
	}
	return err.Success
}

func upload(pf_id, uid, mac string, data []byte, extra string) uint16 {
	key, now := pf_id+"_"+uid, time.Now().Unix()
	switch e := checkMac(pf_id, uid, mac); e {
	case err.Success:
		pSave := &TSaveData{Key: key}
		dbmgo.Find(kDBSave, "_id", key, pSave)
		pSave.CheckSensitiveVal(extra) //敏感数据异动，记下历史存档
		pSave.Data = data
		pSave.UpTime = now
		pSave.Extra = extra
		if dbmgo.DataBase().C(kDBMac).Insert(&MacInfo{mac, key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = now
		}
		dbmgo.UpdateId(kDBSave, pSave.Key, pSave)
		//fmt.Println("---------------upload: ", len(pSave.Data), pSave)
		return err.Success
	case err.Record_cannot_find:
		dbmgo.Insert(kDBSave, &TSaveData{key, data, now, now, 1, extra})
		dbmgo.Insert(kDBMac, &MacInfo{mac, key})
		//fmt.Println("---------------upload new: ", len(pSave.Data), pSave)
		return err.Success
	default:
		return e
	}
}
func download(pf_id, uid, mac string) (*TSaveData, uint16) {
	if errCode := checkMac(pf_id, uid, mac); errCode == err.Success {
		pSave := &TSaveData{Key: pf_id + "_" + uid}
		dbmgo.Find(kDBSave, "_id", pSave.Key, pSave)
		if dbmgo.DataBase().C(kDBMac).Insert(&MacInfo{mac, pSave.Key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = time.Now().Unix()
			dbmgo.UpdateId(kDBSave, pSave.Key, bson.M{"$set": bson.M{"maccnt": pSave.MacCnt, "chtime": pSave.ChTime}})
		}
		return pSave, err.Success
	} else {
		return nil, errCode
	}
}

// ------------------------------------------------------------
// -- Binary 存档
func Rpc_save_upload_binary(req, ack *common.NetPack) { //TODO:zhoumf: 弃用
	args := req.ReadString() //包含多个参数：为了兼容旧客户方~囧
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()
	data := req.LeftBuf()

	//解析组合参数
	list := strings.Split(args, "_")
	length, uid := len(list), ""
	if length > 0 {
		uid = list[0]
	}

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		gamelog.Error("Rpc_save_upload_binary: sign failed")
		ack.WriteUInt16(err.Sign_failed)
		return
	}

	errcode := upload(pf_id, uid, mac, data, "")
	ack.WriteUInt16(errcode)
}
func Rpc_save_upload_binary2(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()
	extra := req.ReadString()
	data := req.ReadLenBuf()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		gamelog.Error("Rpc_save_upload_binary: sign failed")
		ack.WriteUInt16(err.Sign_failed)
		return
	}

	errcode := upload(pf_id, uid, mac, data, extra)
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
		gamelog.Error("Rpc_save_download_binary: sign failed")
		ack.WriteUInt16(err.Sign_failed)
		return
	}

	if ptr, errcode := download(pf_id, uid, mac); errcode == err.Success {
		ack.WriteUInt16(err.Success)
		ack.WriteBuf(ptr.Data)
	} else {
		ack.WriteUInt16(errcode)
	}
}
