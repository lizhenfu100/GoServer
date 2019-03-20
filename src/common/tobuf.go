package common

import (
	"bytes"
	"encoding/gob"
	"unsafe"
)

//【多字符串拼接，用bytes.Buffer.WriteString()快400-500倍】
// bytes.Buffer > string + > fmt.Sprintf > strings.Join

func SwapBuf(a, b *[]byte) { *a, *b = *b, *a }
func ClearBuf(p *[]byte)   { *p = (*p)[:0] } //len(0), cap(old), 旧数据不会修改

// ------------------------------------------------------------
//【仅限只读数据】
func S2B(s string) []byte {
	sh := (*[2]uintptr)(unsafe.Pointer(&s)) //reflect.StringHeader
	bh := [3]uintptr{sh[0], sh[1], sh[1]}   //reflect.SliceHeader
	return *(*[]byte)(unsafe.Pointer(&bh))
}
func B2S(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }

// ------------------------------------------------------------
// go语言间通信用
func T2B(pStruct interface{}) ([]byte, error) { //only public field
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(pStruct); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func B2T(b []byte, pStruct interface{}) error {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	return dec.Decode(pStruct)
}
