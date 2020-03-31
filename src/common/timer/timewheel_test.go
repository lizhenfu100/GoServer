package timer_test

import (
	"common/timer"
	"fmt"
	"testing"
)

// go test -v ./src/common/timer/timewheel_test.go
// TODO：timewheel单元测试

func Test_add(t *testing.T) {
	timer.AddTimer(func() {
		fmt.Println("")
	}, 0, 0, 0)
}
