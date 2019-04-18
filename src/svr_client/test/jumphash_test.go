package test

import (
	"common/std/hash"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// go test -v ./src/svr_client/test/jumphash_test.go

func Test_jumphash(t *testing.T) {
	const kSize = 10000
	rand.Seed(time.Now().Unix())
	list := make([]int, kSize)
	for i := 0; i < kSize; i++ {
		list[i] = rand.Intn(50000)
	}

	out1 := make([]int32, len(list))
	out2 := make([]int32, len(list))
	for k, v := range list {
		out1[k] = hash.JumpHash(uint64(v), 10)
	}
	for k, v := range list {
		out2[k] = hash.JumpHash(uint64(v), 11)
	}
	diff := 0
	for k := range list {
		if out1[k] != out2[k] {
			diff++
		}
	}
	fmt.Println("----------------", float32(diff)/kSize)
}
