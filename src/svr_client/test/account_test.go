package test

import (
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"shared_svr/svr_center/account"
	_ "svr_client/test/init"
	"testing"
)

// go test -v ./src/svr_client/test/account_test.go

// ------------------------------------------------------------
// -- 修复元气测试账号
func Test_account_set_gameinfo(t *testing.T) {
	dbmgo.InitWithUser("52.14.1.205", 27017, "account",
		"chillyroom", "db#233*")
	var list []account.TAccount
	dbmgo.FindAll("Account", bson.M{"gameinfo.SoulKnight": bson.M{"$exists": true}}, &list)
	fmt.Println("-------------------- SoulKnight count:", len(list))
	for i := 0; i < len(list); i++ {
		p := &list[i]
		info := p.GameInfo["SoulKnight"]
		if info.LoginSvrId < 100 {
			info.LoginSvrId += 100
			p.GameInfo["SoulKnight"] = info
			dbmgo.UpdateIdSync("Account", p.AccountID, bson.M{"$set": bson.M{
				"gameinfo.SoulKnight": info}})
		}
	}
}
