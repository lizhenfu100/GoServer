/***********************************************************************
* @ 消息压缩
* @ brief
    1、游戏项目，大部分都是小包
    2、FIXME：snappy/lzf之类的算法，适用于几十字节的无字典压缩
	3、压缩率在几分之一，够用，100左右的也能压到30

* @ author zhoumf
* @ date 2018-4-24
***********************************************************************/
package compress

import (
	"bytes"
	"compress/gzip"
	"conf"
	"encoding/binary"
	"gamelog"
	"io"
	"io/ioutil"
)

const kLimitSize = 128

func CompressTo(b []byte, w io.Writer) {
	if conf.Is_Msg_Compress && len(b) > kLimitSize {
		flag := make([]byte, 4) //前四个字节写压缩标记
		binary.LittleEndian.PutUint32(flag, conf.Flag_Compress)
		w.Write(flag)
		w.Write(Compress(b))
	} else {
		n, e := w.Write(b)
		if n != len(b) || e != nil {
			gamelog.Error("Http ShortWrite: %s", e.Error())
		}
	}
}
func Compress(b []byte) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(b)
	gw.Flush()
	gw.Close()
	return buf.Bytes()
}
func Decompress(b []byte) []byte {
	if len(b) >= 4 && conf.Flag_Compress == binary.LittleEndian.Uint32(b) {
		if gr, err := gzip.NewReader(bytes.NewReader(b[4:])); err == nil {
			ret, err := ioutil.ReadAll(gr)
			gr.Close()
			if err == nil {
				return ret
			}
		}
	}
	return b
}
