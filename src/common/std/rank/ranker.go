/***********************************************************************
* @ 排行榜：数组交换式
* @ brief
	1、从大到小排序，1起始
	2、数组缓存Top N，变动时上下移动
	3、整个排行榜系统包括两部分，本模块负责排序，具体业务模块需单独结构负责存储

* @ author zhoumf
* @ date 2017-9-30
***********************************************************************/
package rank

import (
	"common/assert"
	"dbmgo"
	"sort"
)

type IRankItem interface {
	GetRank() uint
	SetRank(uint)
	Less(IRankItem) bool //内部强转为结构体指针，支持多层级的比对，如：先积分、再等级...
}
type TRanker struct {
	_arr   []IRankItem //0位留空，序号1起始
	_last  uint
	_table string //数据库表名
}

func (self *TRanker) Init(table string, amount int, pRankItemSlice interface{}) {
	self._last = 0
	self._table = table
	self._arr = make([]IRankItem, amount+1)
	dbmgo.FindAll(table, nil, pRankItemSlice)
}

func (self *TRanker) OnValueChange(obj IRankItem) bool {
	oldRank := obj.GetRank()
	newIdx := self.SearchIndex(obj)
	if oldRank > 0 { //已上榜
		self.MoveToIndex(newIdx, oldRank)
	} else if newIdx > 0 { //之前未上榜
		self.InsertToIndex(newIdx, obj)
	}
	return oldRank == obj.GetRank()
}

func (self *TRanker) Clear() {
	for i := uint(0); i <= self._last; i++ {
		self._arr[i] = nil
	}
	self._last = 0
	dbmgo.RemoveAll(self._table, nil)
}
func (self *TRanker) GetRanker(rank uint) IRankItem {
	if rank <= self._last {
		return self._arr[rank]
	}
	return nil
}
func (self *TRanker) GetLastRanker() IRankItem { return self._arr[self._last] }

func (self *TRanker) _WriteDB(obj IRankItem) {
	if dbmgo.InsertSync(self._table, obj) {
	} else {
		dbmgo.UpdateId(self._table, obj.GetRank(), obj)
	}
}

func (self *TRanker) SearchIndex(obj IRankItem) uint {
	//Search()方法回使用“二分查找”算法来搜索某指定切片[0:n]，并返回能够使f(i)=true的最小i；找不到返回n，而不是-1
	idx := sort.Search(len(self._arr), func(i int) bool {
		return i > 0 && (self._arr[i] == nil || self._arr[i].Less(obj))
	})
	if idx == len(self._arr) {
		return 0
	} else {
		return uint(idx)
	}
}
func (self *TRanker) MoveToIndex(dst, src uint) {
	assert.True(dst <= self._last && src <= self._last)
	tmp := self._arr[src]
	if src > dst {
		copy(self._arr[dst+1:], self._arr[dst:src]) //dst后移一步
		self._arr[dst] = tmp
		for i := dst; i <= src; i++ {
			self._arr[i].SetRank(i)
			self._WriteDB(self._arr[i])
		}
	} else if src < dst {
		copy(self._arr[src:dst], self._arr[src+1:]) //src+1前移一步
		self._arr[dst] = tmp
		for i := src; i <= dst; i++ {
			self._arr[i].SetRank(i)
			self._WriteDB(self._arr[i])
		}
	}
}
func (self *TRanker) InsertToIndex(idx uint, obj IRankItem) {
	if idx > self._last {
		self._last++
		idx = self._last
	}
	if self._arr[self._last] != nil {
		self._arr[self._last].SetRank(0) //尾名被挤出
	}
	copy(self._arr[idx+1:], self._arr[idx:self._last])
	self._arr[idx] = obj
	for i := idx; i <= self._last; i++ {
		self._arr[i].SetRank(i)
		self._WriteDB(self._arr[i])
	}
}
