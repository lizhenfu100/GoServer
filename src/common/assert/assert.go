// +build debug

package assert //go run -tags debug main.go

func True(cond bool) {
	if !cond {
		panic("assert")
	}
}
