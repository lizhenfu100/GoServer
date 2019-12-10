package NRIC

import (
	"common"
	"common/std/hash"
	"dbmgo"
	"strconv"
	"time"
)

const KDBTable = "NRIC"

type NRIC struct {
	AidMac   uint32 `bson:"_id"` //账号id 或 设备码hashId
	ID       string //18位身份证
	Name     string
	Birthday int64
	ChTimes  uint8
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
func Rpc_sdk_nric_set(req, ack *common.NetPack) {
	id := req.ReadString()
	name := req.ReadString()
	aid_mac := req.ReadString()
	v := NRIC{AidMac: Parse(aid_mac), ID: id, Name: name}
	if v.AidMac > 0 {
		birthday := GetBirthDay(v.ID)
		if ok, _ := dbmgo.Find(KDBTable, "_id", v.AidMac, &v); ok {
			if birthday > 0 {
				v.ChTimes++
			}
		}
		v.ID = id
		v.Name = name
		v.Birthday = birthday
		if v.Birthday > 0 {
			dbmgo.UpsertId(KDBTable, v.AidMac, &v)
		}
	}
	ack.WriteInt64(v.Birthday)
	ack.WriteString(v.Name)
	ack.WriteByte(v.ChTimes)
}
func Rpc_sdk_nric_birthday(req, ack *common.NetPack) {
	aid_mac := req.ReadString()
	v := NRIC{AidMac: Parse(aid_mac)}
	dbmgo.Find(KDBTable, "_id", v.AidMac, &v)
	ack.WriteInt64(v.Birthday)
	ack.WriteString(v.Name)
	ack.WriteByte(v.ChTimes)
}
