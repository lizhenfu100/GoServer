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
func Shuffle(p []int) {
	for n, i := len(p), 0; i < n; i++ {
		r := rand.Intn(n-i) + i //i前是已洗好的，随机抽后面的，换到i
		p[r], p[i] = p[i], p[r]
	}
}

//生成随机字符串
const (
	_letter = "0123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ"
	idxBit  = 6           //字符池子，只有62个，6位(最大64)够覆盖了
	idxCnt  = 63 / idxBit //63位随机数，可用10次
	idxMask = 1<<idxBit - 1
)

func String(n int) string {
	ret := make([]byte, n)
	for i, r, cnt := n-1, rand.Int63(), idxCnt; i >= 0; {
		if cnt == 0 {
			r, cnt = rand.Int63(), idxCnt
		}
		if idx := int(r & idxMask); idx < len(_letter) {
			ret[i] = _letter[idx]
			i--
		}
		r >>= idxBit
		cnt--
	}
	return common.B2S(ret)
}
