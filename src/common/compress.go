/***********************************************************************
* @ 消息压缩
* @ brief
    1、游戏项目，大部分都是小包
    2、FIXME：snappy/lzf之类的算法，适用于几十字节的无字典压缩
	3、压缩率在几分之一，够用，100左右的也能压到30

* @ author zhoumf
* @ date 2018-4-24
***********************************************************************/
package common

import (
	"bytes"
	"compress/gzip"
	"conf"
	"encoding/binary"
	"io"
	"io/ioutil"
)

const (
	Compress_Limit_Size = 128
	Flag_Compress       = 0x80000000
)

func CompressTo(b []byte, w io.Writer) {
	if conf.Is_Msg_Compress && len(b) > Compress_Limit_Size {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(b)
		gw.Flush()
		gw.Close()
		flag := make([]byte, 4) //前四个字节写压缩标记
		binary.LittleEndian.PutUint32(flag, Flag_Compress)
		w.Write(flag)
		w.Write(buf.Bytes())
	} else {
		w.Write(b)
	}
}
func Compress(b []byte) (ret []byte) {
	if conf.Is_Msg_Compress && len(b) > Compress_Limit_Size {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(b)
		gw.Flush()
		gw.Close()
		flag := make([]byte, 4) //前四个字节写压缩标记
		binary.LittleEndian.PutUint32(flag, Flag_Compress)
		ret = append(ret, flag...)
		ret = append(ret, buf.Bytes()...)
		return ret
	} else {
		return b
	}
}
func Decompress(b []byte) []byte {
	if Flag_Compress == binary.LittleEndian.Uint32(b) {
		if gr, err := gzip.NewReader(bytes.NewReader(b[4:])); err == nil {
			defer gr.Close()
			if ret, err := ioutil.ReadAll(gr); err == nil {
				return ret
			}
		}
	}
	return b
}
