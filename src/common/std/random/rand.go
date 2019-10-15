package random

import (
	"common"
	"math/rand"
)

type Weight interface {
	ID() int
	Weight() int
	SetFlag(bool)
	GetFlag() bool
}

// 按权重随机选出cnt个，不重复
func Select(list []Weight, count int) (ids []int) {
	total, length := 0, len(list)
	for i := 0; i < length; i++ {
		list[i].SetFlag(false)
		total += list[i].Weight()
	}
	if total > 0 {
		for j := 0; j < count; j++ {
			w := rand.Intn(total)
			for i := 0; i < length; i++ {
				p := list[i]
				if p.GetFlag() {
					continue
				}
				if w < p.Weight() {
					ids = append(ids, p.ID())
					p.SetFlag(true)
					total -= p.Weight()
					break
				} else {
					w -= p.Weight()
				}
			}
		}
	}
	return
}

// [left, right]
func Between(left, right int) int { return left + rand.Intn(right+1-left) }

// 数组乱序
func Shuffle(slice []int) {
	length := len(slice)
	for i := 0; i < length; i++ {
		ri := rand.Intn(length-i) + i
		temp := slice[ri]
		slice[i] = temp
		slice[ri] = slice[i]
	}
}

//生成随机字符串s
var g_strBase = []byte("0123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ")

func String(length int) string {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = g_strBase[rand.Intn(len(g_strBase))]
	}
	return common.B2S(ret)
}
