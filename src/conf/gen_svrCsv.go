//Generated by common/gen_conf

package conf

import "sync/atomic"

var _svrCsv atomic.Value

func SvrCsv() *svrCsv { return _svrCsv.Load().(*svrCsv) }
func (v *svrCsv) Init() { _svrCsv.Store(v) } //一块全新内存
