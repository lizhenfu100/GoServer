/***********************************************************************
* @ 云存档
* @ brief
	1、玩家UID、机器码Mac ：一个账号绑定多个设备，一个设备只能绑一个账号
		· 上传下载须查重，该mac是否自己的

	2、机器码放本地存档(加密)中，仅建档时调API取
		· 之后网络交互，从存档取

	3、限制设备更换频率
		· 前三次绑定，可任意时间
		· 后续的绑定，一周一次

	4、云存档里打上玩家标识，比如SaveKey
		· 玩家登录后发现与自己标识不同，禁用
		· 可防止利用云恶意传播

* @ 单机防作弊
	1、后台不断变更密钥，用于金币、钻石、攻击力...敏感数据，防止用户窜改

* @ author zhoumf
* @ date 2018-10-31
***********************************************************************/
package logic

import (
	"common"
	"common/std/sign"
	"dbmgo"
	"fmt"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_save/conf"
	"strings"
	"time"
)

const (
	KDBSave = "Save"
	KDBMac  = "SaveMac"
)

type TSaveData struct {
	Key     string `bson:"_id"` // Pf_id + Uid
	Data    []byte
	UpTime  int64  //上传时刻
	ChTime  int64  //更换时刻
	MacCnt  byte   //绑定的设备数目
	Extra   string //json
	Version string
}
type MacInfo struct {
	Mac string `bson:"_id"` //机器码，取自存档，中途不用API取
	Key string //Pf_id + Uid
}

func GetSaveKey(pf_id, uid string) string { return pf_id + "_" + uid }

func Rpc_save_get_meta_info(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()

	ptr := &TSaveData{Key: GetSaveKey(pf_id, uid)}
	if ok, e := dbmgo.Find(KDBSave, "_id", ptr.Key, ptr); ok {
		ack.WriteInt64(ptr.UpTime)
		ack.WriteInt64(ptr.ChTime)
		ack.WriteUInt8(ptr.MacCnt)
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
	pSave, pMac := &TSaveData{Key: GetSaveKey(pf_id, uid)}, &MacInfo{}
	okMac, _ := dbmgo.Find(KDBMac, "_id", mac, pMac)
	if okMac {
		if pMac.Key != pSave.Key {
			gamelog.Info("Record_mac_already_bind: mac(%s) new(%s) old(%s)", mac, pSave.Key, pMac.Key)
			return err.Record_mac_already_bind
		}
	}
	if okSave, _ := dbmgo.Find(KDBSave, "_id", pSave.Key, pSave); !okSave {
		return err.Record_cannot_find
	}
	if !okMac /*新设备*/ && pSave.MacCnt >= conf.Const.MacFreeBindMax {
		if now := time.Now().Unix(); now-pSave.ChTime < int64(conf.Const.MacChangePeriod) {
			return err.Record_bind_limit
		}
		if pSave.MacCnt >= conf.Const.MacBindMax {
			return err.Record_bind_limit
		}
	}
	return err.Success
}

func upload(pf_id, uid, mac string, data []byte, extra, version string) uint16 {
	key, now := GetSaveKey(pf_id, uid), time.Now().Unix()
	switch e := checkMac(pf_id, uid, mac); e {
	case err.Success:
		pSave := &TSaveData{Key: key}
		dbmgo.Find(KDBSave, "_id", key, pSave)
		if version < pSave.Version { //旧client覆盖新档，报错
			return err.Version_not_match
		}
		pSave.CheckSensitiveVal(extra) //敏感数据异动，记下历史存档
		pSave.Data = data
		pSave.UpTime = now
		pSave.Extra = extra
		pSave.Version = version
		if dbmgo.DataBase().C(KDBMac).Insert(&MacInfo{mac, key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = now
		}
		dbmgo.UpdateId(KDBSave, pSave.Key, pSave)
		//fmt.Println("---------------upload: ", len(pSave.Data), pSave)
		return err.Success
	case err.Record_cannot_find:
		dbmgo.Insert(KDBSave, &TSaveData{key, data, now, now, 1,
			extra, version})
		dbmgo.Insert(KDBMac, &MacInfo{mac, key})
		//fmt.Println("---------------upload new: ", len(pSave.Data), pSave)
		return err.Success
	default:
		return e
	}
}
func download(pf_id, uid, mac, version string) (*TSaveData, uint16) {
	if errCode := checkMac(pf_id, uid, mac); errCode == err.Success {
		pSave := &TSaveData{Key: GetSaveKey(pf_id, uid)}
		dbmgo.Find(KDBSave, "_id", pSave.Key, pSave)
		if version < pSave.Version { //旧client下载新档，报错
			return nil, err.Version_not_match
		}
		if dbmgo.DataBase().C(KDBMac).Insert(&MacInfo{mac, pSave.Key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = time.Now().Unix()
			dbmgo.UpdateId(KDBSave, pSave.Key, bson.M{"$set": bson.M{
				"maccnt": pSave.MacCnt,
				"chtime": pSave.ChTime}})
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

	errcode := upload(pf_id, uid, mac, data, "", "")
	ack.WriteUInt16(errcode)
}
func Rpc_save_upload_binary2(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()
	extra := req.ReadString()
	data := req.ReadLenBuf()
	version := req.ReadString()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		gamelog.Error("Rpc_save_upload_binary: sign failed")
		ack.WriteUInt16(err.Sign_failed)
		return
	}

	errcode := upload(pf_id, uid, mac, data, extra, version)
	ack.WriteUInt16(errcode)
}
func Rpc_save_download_binary(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	Sign := req.ReadString()
	version := req.ReadString()

	//验证签名
	s := fmt.Sprintf("uid=%s&pf_id=%s", uid, pf_id)
	if sign.CalcSign(s) != Sign {
		gamelog.Error("Rpc_save_download_binary: sign failed")
		ack.WriteUInt16(err.Sign_failed)
		return
	}

	if ptr, errcode := download(pf_id, uid, mac, version); errcode == err.Success {
		ack.WriteUInt16(err.Success)
		ack.WriteBuf(ptr.Data)
	} else {
		ack.WriteUInt16(errcode)
	}
}
