package common

import (
	"strconv"
	"strings"
)

type Err string

func (e Err) Error() string { return string(e) }

func InTime(now, begin, end int64) bool {
	return (begin <= 0 || now >= begin) && (end <= 0 || now <= end)
}

// ------------------------------------------------------------
// 跨服服务用PidRpc，需求方可自己组合pid，避免线上项目的数据改造
const KIdMod = 1000

// 头两位预留(可抽一位作角色编号)，三位loginId、四位gameId，最后四位accountId
func PidNew(aid uint32, loginId, gameId int) uint64 {
	return uint64(loginId%KIdMod)<<40 | uint64(gameId%KIdMod)<<32 | uint64(aid)
}
func PidToRoute(pid uint64) (loginId, gameId int) { return int(pid >> 40), int(0xFF & (pid >> 32)) }
func PidToAid(pid uint64) uint32                  { return uint32(pid) }

func IsMatchVersion(a, b string) bool {
	if a == "" || b == "" {
		return true
	}
	// 空版本号能与任意版本匹配
	// 版本号格式：1.12.233，前两组一致的版本间可匹配，第三组用于小调整、bug修复
	idx := strings.LastIndex(a, ".")
	return a[:idx] == b[:idx]
}
func CompareVersion(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	kLen := len(as)
	if len(bs) > kLen {
		kLen = len(bs)
	}
	av, bv := make([]int, kLen), make([]int, kLen)
	for k, v := range as {
		av[k], _ = strconv.Atoi(v)
	}
	for k, v := range bs {
		bv[k], _ = strconv.Atoi(v)
	}
	for i := 0; i < kLen; i++ {
		if av[i] < bv[i] {
			return -1
		} else if av[i] > bv[i] {
			return 1
		}
	}
	return 0
}

// ------------------------------------------------------------
const ( //取数据的几种方式
	Random      = iota
	Core        //集群的核心节点
	ById        //指定取，额外传key
	ModInt      //取模，额外传数字
	ModStr      //取模，额外传字符串
	KeyShuntInt //对谁分流，额外传key、数字
	KeyShuntStr //对谁分流，额外传key、字符串
)

// ------------------------------------------------------------
