//Generated by common/gen_conf

package main

import "sync/atomic"

var _logins atomic.Value

func Logins() logins { return _logins.Load().(logins) }
func (v logins) Init() { _logins.Store(v) } //一块全新内存
