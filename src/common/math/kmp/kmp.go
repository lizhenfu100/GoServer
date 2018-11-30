/***********************************************************************
* @ 模式匹配算法
* @ brief
	S[0]...S[q-k]......S[q-1] S[q]...
			||			||
		   P[0]........P[k-1] P[k]...
		   		 ||		||
		   		P[0]...P[j-1] P[j]...P[k]...

	P[k]已经失配了，之前的子串又是相同的
	我们要找个同样也是P[0]打头的子串即P[0]···P[j-1](j==next[k-1])，看看它的下一项P[j]是否能匹配
	P[j-1]...P[0] 与 P[k-1]...P[k-j] 相同，失配后的游标回置到[j]即可

* @ author zhoumf
* @ date 2018-11-30
***********************************************************************/
package kmp

import (
	"errors"
	"fmt"
)

// KMP字符串匹配，用于脏字库排查
type kmp struct {
	pattern string
	next    []int
	size    int
}

func NewKMP(pattern string) (*kmp, error) {
	if next, err := makeNext(pattern); err != nil {
		return nil, err
	} else {
		return &kmp{
			pattern: pattern,
			next:    next,
			size:    len(pattern)}, nil
	}
}
func (self *kmp) String() string { //For debugging
	return fmt.Sprintf("pattern: %v\nnext: %v", self.pattern, self.next)
}

// returns an array containing indexes of matches
// - error if pattern argument is less than 1 char
func makeNext(pattern string) ([]int, error) {
	// sanity check
	length := len(pattern)
	if length == 0 {
		return nil, errors.New("'pattern' must contain at least one character")
	}
	if length == 1 {
		return []int{-1}, nil
	}
	next := make([]int, length)
	next[0], next[1] = -1, 0

	pos, count := 2, 0
	for pos < length {
		if pattern[pos-1] == pattern[count] {
			count++
			next[pos] = count
			pos++
		} else {
			if count > 0 {
				count = next[count]
			} else {
				next[pos] = 0
				pos++
			}
		}
	}
	return next, nil
}

// return index of first occurence of kmp.pattern in argument 's'
// - if not found, returns -1
func (self *kmp) FindStringIndex(s string) int {
	// sanity check
	if len(s) < self.size {
		return -1
	}
	m, i := 0, 0
	for m+i < len(s) {
		if self.pattern[i] == s[m+i] {
			if i == self.size-1 {
				return m
			}
			i++
		} else {
			m += i - self.next[i]
			if self.next[i] > -1 {
				i = self.next[i]
			} else {
				i = 0
			}
		}
	}
	return -1
}

const startSize = 10 //for effeciency, define default array-size

// find every occurence of the kmp.pattern in 's'
func (self *kmp) FindAllStringIndex(s string) []int {
	// precompute
	len_s := len(s)
	if len_s < self.size {
		return []int{}
	}

	match := make([]int, 0, startSize)
	m, i := 0, 0
	for m+i < len_s {
		if self.pattern[i] == s[m+i] {
			if i == self.size-1 {
				// the word was matched
				match = append(match, m)
				// simulate miss, and keep running
				m += i - self.next[i]
				if self.next[i] > -1 {
					i = self.next[i]
				} else {
					i = 0
				}
			} else {
				i++
			}
		} else {
			m += i - self.next[i]
			if self.next[i] > -1 {
				i = self.next[i]
			} else {
				i = 0
			}
		}
	}
	return match
}
