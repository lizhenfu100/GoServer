/***********************************************************************
* @ 云存档
* @ brief
	1、玩家UID、机器码Mac ：一个账号绑定多个设备，一个设备只能绑一个账号
		· 上传下载须查重，该mac是否自己的

	2、机器码放本地存档(加密)中，仅建档时调API取
		· 之后网络交互，从存档取

	3、限制设备更换频率
		· 前几次绑定，可任意时间
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
	"dbmgo"
	"gamelog"
	"generate_out/err"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_save/conf"
	"time"
)

const (
	KDBSave = "Save"
	KDBMac  = "SaveMac"
)

type TSaveData struct {
	Key       string `bson:"_id"` // Pf_id + Uid
	Data      []byte `json:"-"`
	UpTime    int64  //上传时刻
	ChTime    int64  //更换时刻
	RaiseTime int64  //提升绑定上限的时刻
	MacCnt    byte   //绑定的设备次数
	Extra     string //json TSensitive
	Version   string
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

	_, errCode := checkMac(pf_id, uid, mac)
	ack.WriteUInt16(errCode)
}
func checkMac(pf_id, uid, mac string) (*TSaveData, uint16) { //Notice：不可调换错误码判断顺序
	pSave, pMac := &TSaveData{Key: GetSaveKey(pf_id, uid)}, &MacInfo{}
	oldMac, _ := dbmgo.Find(KDBMac, "_id", mac, pMac)
	if oldMac && pMac.Key != pSave.Key {
		gamelog.Info("Record_mac_already_bind: mac(%s) new(%s) old(%s)", mac, pSave.Key, pMac.Key)
		return pSave, err.Record_mac_already_bind
	}
	if ok, _ := dbmgo.Find(KDBSave, "_id", pSave.Key, pSave); !ok {
		gamelog.Track("Record_cannot_find: key(%s)", pSave.Key)
		return pSave, err.Record_cannot_find
	}
	if !oldMac /*新设备*/ && pSave.MacCnt >= conf.Const.MacFreeBindMax {
		now := time.Now().Unix()
		if now-pSave.ChTime < int64(conf.Const.MacChangePeriod) {
			gamelog.Track("Record_bind_limit: %v", pSave)
			return pSave, err.Record_bind_limit
		}
		if pSave.MacCnt >= conf.Const.MacBindMax {
			if (now-pSave.RaiseTime)/(3600*24) < int64(conf.Const.RaiseBindCntDay) {
				gamelog.Track("Record_bind_max: %v", pSave)
				return pSave, err.Record_bind_max
			} else {
				pSave.MacCnt-- //90天，绑定次数+1
				pSave.RaiseTime = now
				return pSave, err.Success
			}
		}
	}
	return pSave, err.Success
}

func upload(pf_id, uid, mac string, data []byte, extra, clientVersion string) uint16 {
	now := time.Now().Unix()
	pSave, errCode := checkMac(pf_id, uid, mac)
	switch errCode {
	case err.Success:
		if clientVersion < pSave.Version { //旧client覆盖新档，报错
			return err.Version_not_match
		}
		pSave.CheckBackup(extra) //敏感数据异动，记下历史存档
		pSave.Data = data
		pSave.UpTime = now
		pSave.Extra = extra
		pSave.Version = clientVersion
		if dbmgo.DB().C(KDBMac).Insert(&MacInfo{mac, pSave.Key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = now
		}
		dbmgo.UpdateId(KDBSave, pSave.Key, pSave)
		//fmt.Println("---------------upload: ", len(pSave.Data), pSave)
		return err.Success
	case err.Record_cannot_find:
		pSave.Data = data
		pSave.UpTime = now
		pSave.ChTime = now
		pSave.RaiseTime = 0
		pSave.MacCnt = 1
		pSave.Extra = extra
		pSave.Version = clientVersion
		dbmgo.Insert(KDBSave, pSave)
		dbmgo.Insert(KDBMac, &MacInfo{mac, pSave.Key})
		//fmt.Println("---------------upload new: ", len(pSave.Data), pSave)
		return err.Success
	default:
		return errCode
	}
}
func download(pf_id, uid, mac, clientVersion string) (*TSaveData, uint16) {
	if pSave, errCode := checkMac(pf_id, uid, mac); errCode == err.Success {
		if clientVersion < pSave.Version { //旧client下载新档，报错
			return nil, err.Version_not_match
		}
		if dbmgo.DB().C(KDBMac).Insert(&MacInfo{mac, pSave.Key}) == nil {
			pSave.MacCnt++
			pSave.ChTime = time.Now().Unix()
			dbmgo.UpdateId(KDBSave, pSave.Key, bson.M{"$set": bson.M{
				"maccnt":    pSave.MacCnt,
				"chtime":    pSave.ChTime,
				"raisetime": pSave.RaiseTime}})
		}
		return pSave, err.Success
	} else {
		return nil, errCode
	}
}

// ------------------------------------------------------------
// -- Binary 存档
func Rpc_save_upload_binary2(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	req.ReadString() //sign签名：不必要的，底层对消息加密
	extra := req.ReadString()
	//存档
	cnt := req.ReadUInt16()
	old := req.ReadPos
	req.ReadPos += int(cnt)
	data := req.Data()[old:req.ReadPos]

	clientVersion := req.ReadString()

	errcode := upload(pf_id, uid, mac, data, extra, clientVersion)
	ack.WriteUInt16(errcode)
}
func Rpc_save_upload_binary(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	req.ReadString() //sign签名：不必要的，底层对消息加密
	extra := req.ReadString()
	data := req.ReadLenBuf()
	clientVersion := req.ReadString()

	errcode := upload(pf_id, uid, mac, data, extra, clientVersion)
	ack.WriteUInt16(errcode)
}
func Rpc_save_download_binary(req, ack *common.NetPack) {
	uid := req.ReadString()
	pf_id := req.ReadString()
	mac := req.ReadString()
	req.ReadString() //sign签名：不必要的，底层对消息加密
	clientVersion := req.ReadString()

	if ptr, errcode := download(pf_id, uid, mac, clientVersion); errcode == err.Success {
		ack.WriteUInt16(err.Success)
		ack.WriteBuf(ptr.Data)
	} else {
		ack.WriteUInt16(errcode)
	}
}
