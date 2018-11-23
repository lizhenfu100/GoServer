package kmp

import (
	"errors"
	"fmt"
)

// KMP字符串匹配，用于脏字库排查
type kmp struct {
	pattern string
	prefix  []int
	size    int
}

func NewKMP(pattern string) (*kmp, error) {
	prefix, err := computePrefix(pattern)
	if err != nil {
		return nil, err
	}
	return &kmp{
			pattern: pattern,
			prefix:  prefix,
			size:    len(pattern)},
		nil
}
func (self *kmp) String() string { //For debugging
	return fmt.Sprintf("pattern: %v\nprefix: %v", self.pattern, self.prefix)
}

// returns an array containing indexes of matches
// - error if pattern argument is less than 1 char
func computePrefix(pattern string) ([]int, error) {
	// sanity check
	len_p := len(pattern)
	if len_p < 2 {
		if len_p == 0 {
			return nil, errors.New("'pattern' must contain at least one character")
		}
		return []int{-1}, nil
	}
	t := make([]int, len_p)
	t[0], t[1] = -1, 0

	pos, count := 2, 0
	for pos < len_p {
		if pattern[pos-1] == pattern[count] {
			count++
			t[pos] = count
			pos++
		} else {
			if count > 0 {
				count = t[count]
			} else {
				t[pos] = 0
				pos++
			}
		}
	}
	return t, nil
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
			m = m + i - self.prefix[i]
			if self.prefix[i] > -1 {
				i = self.prefix[i]
			} else {
				i = 0
			}
		}
	}
	return -1
}

const startSize = 10 //for effeciency, define default array-size

// find every occurence of the kmp.pattern in 's'
func (kmp *kmp) FindAllStringIndex(s string) []int {
	// precompute
	len_s := len(s)
	if len_s < kmp.size {
		return []int{}
	}

	match := make([]int, 0, startSize)
	m, i := 0, 0
	for m+i < len_s {
		if kmp.pattern[i] == s[m+i] {
			if i == kmp.size-1 {
				// the word was matched
				match = append(match, m)
				// simulate miss, and keep running
				m = m + i - kmp.prefix[i]
				if kmp.prefix[i] > -1 {
					i = kmp.prefix[i]
				} else {
					i = 0
				}
			} else {
				i++
			}
		} else {
			m = m + i - kmp.prefix[i]
			if kmp.prefix[i] > -1 {
				i = kmp.prefix[i]
			} else {
				i = 0
			}
		}
	}
	return match
}
