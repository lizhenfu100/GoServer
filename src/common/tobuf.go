package common

import (
	"bytes"
	"encoding/gob"
	// "encoding/json"
)

// stData := std.IntPair{11, 22}
// b, _ := json.Marshal(&stData)
// var data std.IntPair
// json.Unmarshal(b, &data)
// fmt.Println(data)

func ToBytes(pStruct interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(pStruct)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func ToStruct(b []byte, pStruct interface{}) error {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	return dec.Decode(pStruct)
}

func SwapBuf(rhs, lhs *[]byte) {
	temp := *rhs
	*rhs = *lhs
	*lhs = temp
}
func ClearBuf(p *[]byte) {
	// len(0), cap(old), 旧数据不会修改
	// *p = append((*p)[:0], []byte{}...)
	*p = (*p)[:0]
}
