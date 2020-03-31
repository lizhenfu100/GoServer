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
type Empty struct{} //unsafe.Sizeof(Empty) //0
type Set map[interface{}]Empty

// ------------------------------------------------------------
// 数组封装
type (
	Ints    []int
	Int64s  []int64
	UInt32s []uint32
	UInt64s []uint64
	Strings []string
)

func (p *Ints) Add(v int)       { *p = append(*p, v) }
func (p *Int64s) Add(v int64)   { *p = append(*p, v) }
func (p *UInt32s) Add(v uint32) { *p = append(*p, v) }
func (p *UInt64s) Add(v uint64) { *p = append(*p, v) }
func (p *Strings) Add(v string) { *p = append(*p, v) }

func (p *Ints) Del(i int)            { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x Ints) Less(i, j int) bool    { return x[i] < x[j] }
func (x Ints) Swap(i, j int)         { x[i], x[j] = x[j], x[i] }
func (p *Int64s) Del(i int)          { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x Int64s) Less(i, j int) bool  { return x[i] < x[j] }
func (x Int64s) Swap(i, j int)       { x[i], x[j] = x[j], x[i] }
func (p *UInt32s) Del(i int)         { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x UInt32s) Less(i, j int) bool { return x[i] < x[j] }
func (x UInt32s) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (p *UInt64s) Del(i int)         { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x UInt64s) Less(i, j int) bool { return x[i] < x[j] }
func (x UInt64s) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (p *Strings) Del(i int)         { *p = append((*p)[:i], (*p)[i+1:]...) }
func (x Strings) Less(i, j int) bool { return x[i] < x[j] }
func (x Strings) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

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
func (x Strings) Index(v string) int {
	for i := 0; i < len(x); i++ {
		if v == x[i] {
			return i
		}
	}
	return -1
}
func (x UInt64s) Index(v uint64) int {
	for i := 0; i < len(x); i++ {
		if v == x[i] {
			return i
		}
	}
	return -1
}
func (x Int64s) Index(v int64) int {
	for i := 0; i < len(x); i++ {
		if v == x[i] {
			return i
		}
	}
	return -1
}

// ------------------------------------------------------------
//
