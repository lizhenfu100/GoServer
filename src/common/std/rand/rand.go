package rand

import (
	"common"
	"math/rand"
	"time"
)

type Weight interface {
	ID() int
	Weight() int
	SetFlag(bool)
	GetFlag() bool
}

// 按权重随机选出cnt个，不重复
func RandSelect(list []Weight, count int) (ids []int) {
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
func RandBetween(left, right int) int { return left + rand.Intn(right+1-left) }

// 数组乱序
func RandShuffle(slice []int) {
	length := len(slice)
	for i := 0; i < length; i++ {
		ri := rand.Intn(length-i) + i
		temp := slice[ri]
		slice[i] = temp
		slice[ri] = slice[i]
	}
}

//生成随机字符串s
var g_strBase = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(length int) string {
	result := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result[i] = g_strBase[r.Intn(len(g_strBase))]
	}
	return common.ToStr(result)
}
