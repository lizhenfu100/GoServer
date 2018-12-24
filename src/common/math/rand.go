package math

import (
	"math/rand"
	"time"
)

type RandItem struct {
	ID         int
	Weight     int
	isSelected bool
}

// 按权重随机选出cnt个，不重复
func RandSelect(list []RandItem, count int) (ret []int) {
	total, length := 0, len(list)
	// for _, v := range list {
	// 	total += v.Weight
	// 	v.isSelected = false //【Bug】range值拷贝迭代的坑啊
	// }
	for i := 0; i < length; i++ {
		list[i].isSelected = false
		total += list[i].Weight
	}

	if count > length {
		return nil
	}

	for j := 0; j < count; j++ {
		rand := rand.Intn(total)
		for i := 0; i < length; i++ {
			data := &list[i]
			if data.isSelected {
				continue
			}
			if rand < data.Weight {
				ret = append(ret, data.ID)
				data.isSelected = true
				total -= data.Weight
				break
			} else {
				rand -= data.Weight
			}
		}
	}
	return ret
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
	return string(result)
}
