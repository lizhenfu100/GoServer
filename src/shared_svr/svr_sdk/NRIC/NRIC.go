/***********************************************************************
Rpc_nric_set uint16 = 162
    上行：string 身份证号, string 姓名, string 账号id或设备码
    下行：uint6 错误码，int64 生日时间戳，byte 修改次数，uint32 身份信息哈希值
    err.Name_format_err    名字不是中文
    err.Invalid            身份证不合法，账号或设备码是空的
    err.Change_times_max   修改次数已达上限


// 获取生日时间戳
Rpc_nric_birthday uint16 = 163
    上行：string 账号id或设备码
    下行：uint6 错误码，int64 生日时间戳，byte 修改次数，uint32 身份信息哈希值
    err.Not_found            无此信息
    err.Unknow_error         网络错误，db错误

// 拷贝实名信息到其它账号或设备
Rpc_nric_copy_to uint16 = 166
    上行：int32 身份信息哈希值，string 账号id或设备码
    下行：uint6 错误码，int64 生日时间戳，byte 修改次数
    err.Not_found
    err.Unknow_error
    err.Invalid			账号或设备码传了空值

// 辅助单机游戏记录在线时长
Rpc_nric_online_time_set uint16 = 165
    上行：string 设备码，int32，int32

Rpc_nric_online_time_get uint16 = 164
    上行：string 设备码
    下行：uint6 错误码，回传set的两个int值
    err.Not_found
    err.Unknow_error
    err.Invalid			账号或设备码传了空值
***********************************************************************/
package NRIC

import (
	"common"
	"common/std/hash"
	"common/std/sign/aes"
	"dbmgo"
	"generate_out/err"
	"strconv"
	"strings"
	"time"
)

const KDBTable = "NRIC"

type NRIC struct {
	AidMac     uint32 `bson:"_id"` //账号id 或 设备码hashId
	ID         []byte //18位身份证，加密
	Name       []byte //姓名，加密
	Birthday   int64
	PersonHash uint32 //hash(身份证号+姓名)，玩家身份信息最终不会保留
	ChTimes    uint8
}

var wi = []byte{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var mi = []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

func GetBirthDay(cardId string) int64 {
	if len(cardId) == 18 {
		sum := 0
		for i := 0; i < 17; i++ {
			sum += int((cardId[i] - '0') * wi[i])
		}
		if mi[sum%11] == cardId[17] {
			return birthDay(cardId)
		}
	}
	return 0
}
func birthDay(id string) int64 {
	year1 := id[6:10]
	month1 := id[10:12]
	day1 := id[12:14]
	year, _ := strconv.Atoi(year1)
	month, _ := strconv.Atoi(month1)
	day, _ := strconv.Atoi(day1)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Unix()
}

func Parse(aid_mac string) uint32 {
	if aid, e := strconv.Atoi(aid_mac); e == nil {
		return uint32(aid)
	} else {
		return hash.StrHash(aid_mac)
	}
}

// ------------------------------------------------------------
func Rpc_nric_set(req, ack *common.NetPack, _ common.Conn) {
	id := strings.ToUpper(req.ReadString())
	name := req.ReadString() //不检查名字格式，少数民族的各种奇葩
	aid_mac := req.ReadString()
	v := NRIC{
		AidMac:     Parse(aid_mac),
		ID:         aes.Encode(id),
		Name:       aes.Encode(name),
		Birthday:   GetBirthDay(id),
		PersonHash: hash.StrHash(id + name),
	}
	if v.AidMac <= 0 || v.Birthday == 0 {
		ack.WriteUInt16(err.Invalid)
	} else {
		old := NRIC{}
		if ok, _ := dbmgo.Find(KDBTable, "_id", v.AidMac, &old); ok {
			if v.ChTimes = old.ChTimes + 1; v.ChTimes > 1 {
				ack.WriteUInt16(err.Change_times_max)
				return
			}
		}
		dbmgo.UpsertId(KDBTable, v.AidMac, &v)
		ack.WriteUInt16(err.Success)
		ack.WriteInt64(v.Birthday)
		ack.WriteByte(v.ChTimes)
		ack.WriteUInt32(v.PersonHash)
		ack.WriteString(name)
	}
}
func Rpc_nric_birthday(req, ack *common.NetPack, _ common.Conn) {
	aid_mac := req.ReadString()
	v := NRIC{AidMac: Parse(aid_mac)}
	if ok, e := dbmgo.Find(KDBTable, "_id", v.AidMac, &v); ok {
		ack.WriteUInt16(err.Success)
		ack.WriteInt64(v.Birthday)
		ack.WriteByte(v.ChTimes)
		ack.WriteUInt32(v.PersonHash)
		ack.WriteString(common.B2S(aes.Decode(v.Name)))
	} else if e == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		ack.WriteUInt16(err.Unknow_error)
	}
}
func Rpc_nric_copy_to(req, ack *common.NetPack, _ common.Conn) {
	personHash := req.ReadUInt32()
	aid_mac := req.ReadString()
	v := NRIC{}
	if aidmac := Parse(aid_mac); aidmac <= 0 {
		ack.WriteUInt16(err.Invalid)
	} else if ok, e := dbmgo.Find(KDBTable, "personhash", personHash, &v); ok {
		v.AidMac = aidmac
		dbmgo.UpsertId(KDBTable, v.AidMac, &v)
		ack.WriteUInt16(err.Success)
		ack.WriteInt64(v.Birthday)
		ack.WriteByte(v.ChTimes)
		ack.WriteString(common.B2S(aes.Decode(v.Name)))
	} else if e == nil {
		ack.WriteUInt16(err.Not_found)
	} else {
		ack.WriteUInt16(err.Unknow_error)
	}
}
