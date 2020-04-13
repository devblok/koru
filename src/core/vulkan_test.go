// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core_test

import (
	"testing"

	"github.com/devblok/koru/src/core"
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
