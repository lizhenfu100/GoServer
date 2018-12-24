/***********************************************************************
* @ 入库的全局参数
* @ brief
	、本模块的接口性能不高（数据库同步操作）
	、业务模块管理各自的参数缓存

* @ author zhoumf
* @ date 2018-12-7
***********************************************************************/
package dbmgo

const KDBSvrArgs = "SvrArgs"

//type IArgs interface {
//	ReadDB() bool //return Find(dbmgo.KDBSvrArgs, "_id", key, pVal)
//	UpdateDB()    //UpdateId(dbmgo.KDBSvrArgs, key, pVal)
//	InsertDB()    //Insert(dbmgo.KDBSvrArgs, pVal)
//}
