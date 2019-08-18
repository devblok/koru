package core_test

import (
	"testing"

	"github.com/devblok/koru/core"
)

func BenchmarkSliceUint32Small(b *testing.B) {
	data := make([]byte, 100)
	for idx := 0; idx < b.N; idx++ {
		core.SliceUint32(data)
	}
}

func BenchmarkSliceUint32Medium(b *testing.B) {
	data := make([]byte, 1000)
	for idx := 0; idx < b.N; idx++ {
		core.SliceUint32(data)
	}
}

func BenchmarkSliceUint32Big(b *testing.B) {
	data := make([]byte, 100000)
	for idx := 0; idx < b.N; idx++ {
		core.SliceUint32(data)
	}
}
