/***********************************************************************
* @ 入库的全局参数
* @ brief
	、本模块的接口性能不高（数据库同步操作）
	、业务模块管理各自的参数缓存

* @ author zhoumf
* @ date 2018-12-7
***********************************************************************/
package dbmgo

const KTableArgs = "args"

//type IArgs interface {
//	ReadDB() bool //return Find(dbmgo.KTableArgs, "_id", DBKey, pVal)
//	UpdateDB()    //UpdateId(dbmgo.KTableArgs, DBKey, pVal)
//	InsertDB()    //Insert(dbmgo.KTableArgs, pVal)
//	InitDB()      // if !Find() { Insert() }
//}
