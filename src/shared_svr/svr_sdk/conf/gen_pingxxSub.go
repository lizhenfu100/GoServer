//Generated by common/gen_conf

package conf

import "sync/atomic"

var _pingxxSub atomic.Value

func PingxxSub() pingxxSub { return _pingxxSub.Load().(pingxxSub) }
func NilPingxxSub() pingxxSub { return nil }
func (v pingxxSub) Init() { _pingxxSub.Store(v) } //一块全新内存
