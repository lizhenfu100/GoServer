// +build debug

package assert //go run -tags debug

func True(cond bool) {
	if !cond {
		panic("assert")
	}
}
