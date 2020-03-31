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
package std

// KMP字符串匹配，用于脏字库排查
type kmp struct {
	pattern string
	next    []int
}

func NewKMP(pattern string) *kmp {
	if next := makeNext(pattern); next != nil {
		return &kmp{pattern, next}
	}
	return nil

}

// returns an array containing indexes of matches
func makeNext(pattern string) []int {
	length := len(pattern)
	if length == 0 {
		return nil
	}
	if length == 1 {
		return []int{-1}
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
	return next
}

// return index of first occurence of kmp.pattern in argument 's'
// - if not found, returns -1
func (self *kmp) FindStringIndex(s string) int {
	if size, cnt := len(self.pattern), len(s); cnt >= size {
		for m, i := 0, 0; m+i < cnt; {
			if self.pattern[i] == s[m+i] {
				if i == size-1 {
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
	}
	return -1
}

const startSize = 10 //for effeciency, define default array-size

// find every occurence of the kmp.pattern in 's'
func (self *kmp) FindAllStringIndex(s string) []int {
	if size, cnt := len(self.pattern), len(s); cnt < size {
		return []int{}
	} else {
		match := make([]int, 0, startSize)
		for m, i := 0, 0; m+i < cnt; {
			if self.pattern[i] == s[m+i] {
				if i == size-1 {
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
}
