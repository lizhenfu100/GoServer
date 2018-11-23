package unit_test

import (
	"common"
	"fmt"
	"testing"
)

type TTestMap struct {
	key string
	val int
}

func Test_List(t *testing.T) {
	var list []int
	fmt.Println(len(list)) // 0
	if list == nil {       //! 判断通过哟！
		fmt.Println("var list []int: ", list) // []
	}
	if make([]byte, 0) == nil {
		fmt.Println("make([]byte, 0) is nil !!!")
	} else {
		fmt.Println("make([]byte, 0) isn't nil !!!")
	}
	list = append(list, 22)
	list = append(list, 33)
	fmt.Println(list, list[2:]) // [22,33] []  可以填数组长度哟！！

	for i := 0; i < len(list); i++ {
		if list[i] == 22 {
			list = append(list[:i], list[i+1:]...)
			i--
		}
	}
	fmt.Println(list) // []

	buf := make([]byte, 4)
	buf1 := buf[1:3]
	buf1[0] = 5
	fmt.Println(buf)
	fmt.Println(buf1, len(buf1))
	fmt.Println(&buf[0], &buf[1], &buf[2])
	fmt.Println(&buf1[0], &buf1[1])

	common.ClearBuf(&buf)
	fmt.Println("--Clear--")
	fmt.Println(buf, len(buf), cap(buf))
	fmt.Println(buf1, len(buf1), cap(buf1))
	buf = append(buf, []byte{1, 2, 3}...)
	fmt.Println(buf, &buf[0], &buf[1])
	fmt.Println(buf1, &buf1[0], &buf1[1])
}
func Test_Map(t *testing.T) {
	dict := make(map[int]TTestMap)
	// dict := make(map[int]*TTestMap)
	dict[1] = TTestMap{"aa", 11}
	dict[2] = TTestMap{"bb", 22}
	dict[3] = TTestMap{"cc", 33}

	v, _ := dict[1]
	v.key = "zhoumf" //! map取出来是拷贝，原结构里的值没变
	fmt.Println(dict)
}
