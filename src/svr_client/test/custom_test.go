package test

import (
	"fmt"
	_ "svr_client/test/init"
	"sync"
	"testing"
)

func Test_1(t *testing.T) {
	var v interface{} = uint32(10)
	v2 := uint32(10)
	fmt.Println(v == v2)

	var map1 sync.Map
	map1.Store("a", int(10))
	vv, _ := map1.Load("a")
	fmt.Println(vv == t)
}

func Test_2(t *testing.T) {
	lst := new([]int)
	lst1 := lst
	lst2 := lst
	*lst = append(*lst, 23)
	fmt.Println(lst, lst1, lst2)

	team := *lst
	*lst = append(team, 1)
	fmt.Println(lst, lst1, lst2)
}
