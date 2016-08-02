package common

import (
	"bytes"
	"encoding/gob"
)

func ToBytes(Struct interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(Struct)
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
