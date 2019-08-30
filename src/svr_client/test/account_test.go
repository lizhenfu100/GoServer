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
// -- 修改账号
func Test_account_set_bindinfo(t *testing.T) {
	dbmgo.InitWithUser("3.17.67.102", 27017, "account",
		"chillyroom", "db#233*")
	coll := dbmgo.DB().C("Account")

	var list []account.TAccount
	//dbmgo.FindAll(account.KDBTable, bson.M{}, &list)
	//for _, v := range list {
	//	fmt.Println(v.AccountID, v.Name)
	//}

	//bindVal := "1003303623@qq.com"
	q := coll.Find(bson.M{
		"bindinfo.email": bson.M{"$exists": false},
		"bindinfo.name":  bson.M{"$exists": false},
	})
	if err := q.All(&list); err == nil {
		//fmt.Println("-------- ok", n)
		for _, v := range list {
			fmt.Println("---------------- ", v.Name, v.AccountID)
		}
	} else {
		fmt.Println("-------- err: ", err.Error())
	}
	fmt.Println("... finish ...")
}
