package common

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"io/ioutil"
)

const (
	Compress_Limit_Size = 128
	Compress_Flag       = 0x80000000
)

func CompressInto(b []byte, w io.Writer) {
	if len(b) < Compress_Limit_Size { //不压缩
		w.Write(b)
	} else {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(b)
		gw.Flush()
		gw.Close()
		flag := make([]byte, 4) //前四个字节写压缩标记
		binary.LittleEndian.PutUint32(flag, Compress_Flag)
		w.Write(flag)
		w.Write(buf.Bytes())
	}
}
func Decompress(b []byte) []byte {
	if Compress_Flag == binary.LittleEndian.Uint32(b) {
		if gr, err := gzip.NewReader(bytes.NewReader(b[4:])); err == nil {
			if b2, err := ioutil.ReadAll(gr); err == nil {
				return b2
			}
			gr.Close()
		}
	}
	return b
}
