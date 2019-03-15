package std

type IntPair struct {
	ID  int
	Cnt int
}
type KeyPair struct {
	Name string
	ID   int
}
type StrPair struct {
	K string
	V string
}
type Addr struct {
	IP   string
	Port uint16
}

// ------------------------------------------------------------
//! 数组封装
type (
	Ints    []int
	UInt32s []uint32
)

func (p *Ints) Add(v int)               { *p = append(*p, v) }
func (p *Ints) Del(i int)               { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x Ints) Less(i, j int) bool       { return x[i] < x[j] }
func (x Ints) Swap(i, j int)            { x[i], x[j] = x[j], x[i] }
func (p *UInt32s) Add(v uint32)         { *p = append(*p, v) }
func (p *UInt32s) Del(i int)            { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x UInt32s) Less(i, j uint32) bool { return x[i] < x[j] }
func (x UInt32s) Swap(i, j uint32)      { x[i], x[j] = x[j], x[i] }
func (x UInt32s) Index(v uint32) int {
	for i := 0; i < len(x); i++ {
		if v == x[i] {
			return i
		}
	}
	return -1
}
func (x Ints) Index(v int) int {
	for i := 0; i < len(x); i++ {
		if v == x[i] {
			return i
		}
	}
	return -1
}

// ------------------------------------------------------------
//
