/***********************************************************************
* @ 玩家数据升级
* @ brief
	1、玩家登录时，先从数据库读版本号，不匹配则升级成对应版本

	2、version目录下存储历史版本的玩家数据结构（历史代码copy到版本号目录）
		、玩家数据是离散式的，可以只存变动过的那一部分，针对该部分更新

	3、筛选升级须经过的版本(可能连续多个)，依次调用升级API

	4、完毕后，库数据应同当前节点一致

* @ author zhoumf
* @ date 2019-1-4
***********************************************************************/
package version

import (
//v0 "svr_game/version/0"
//v1_1 "svr_game/version/0.1.1"
//v1_2 "svr_game/version/0.1.2"
//v2_1 "svr_game/version/0.2.1"
)

func Upgrade(pid uint32, oldVersion, newVersion string) {
	//1、找到 [old, new) 两者之间的版本目录
	oldIdx, newIdx := 0, 0
	for i, v := range g_api {
		if oldVersion >= v.version {
			oldIdx = i
		} else if newVersion >= v.version {
			newIdx = i
		}
	}
	//2、依次执行升级接口（不包括new版本）
	for i := oldIdx; i < newIdx; i++ {
		g_api[i].api(pid)
	}
}

type verApi struct {
	version string
	api     func(uint32)
}

var g_api = []verApi{
//{"", v0.Upgrade},
//{"0.1.1", v1_1.Upgrade}, //例子
//{"0.1.2", v1_2.Upgrade},
//{"0.2.1", v2_1.Upgrade},
}
