package safe

import "sync"

type Strings struct {
	sync.Mutex //no copy
	v          []string
}

func (p *Strings) Add(v string) {
	p.Lock()
	p.v = append(p.v, v)
	p.Unlock()
}
func (p *Strings) Del(i int) {
	p.Lock()
	p.v = append(p.v[:i], p.v[i+1:]...)
	p.Unlock()
}
func (p *Strings) Less(i, j int) bool {
	p.Lock()
	ret := p.v[i] < p.v[j]
	p.Unlock()
	return ret
}
func (p *Strings) Swap(i, j int) {
	p.Lock()
	p.v[i], p.v[j] = p.v[j], p.v[i]
	p.Unlock()
}
func (p *Strings) Index(v string) (ret int) {
	ret = -1
	p.Lock()
	for i := 0; i < len(p.v); i++ {
		if v == p.v[i] {
			ret = i
		}
	}
	p.Unlock()
	return
}
func (p *Strings) MoveTo(ret *[]string) {
	p.Lock()
	for _, v := range p.v {
		*ret = append(*ret, v)
	}
	p.v = p.v[:0]
	p.Unlock()
}
func (p *Strings) CopyTo(ret *[]string) {
	p.Lock()
	for _, v := range p.v {
		*ret = append(*ret, v)
	}
	p.Unlock()
}
func (p *Strings) Size() int {
	p.Lock()
	ret := len(p.v)
	p.Unlock()
	return ret
}
