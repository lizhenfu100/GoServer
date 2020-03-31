// +build debug

package assert //go run -tags debug

const IsDebug = true

func True(cond bool) {
	if !cond {
		panic("assert")
	}
}
