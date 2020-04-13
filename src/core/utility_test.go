// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/devblok/src/koru/core"
	"github.com/gobuffalo/packr"
)

var (
	StaticResources packr.Box
	testImage       image.Image
)

func init() {
	StaticResources = packr.NewBox("../assets")
	img, err := png.Decode(bytes.NewReader(StaticResources.Bytes("Bricks_COLOR.png")))
	if err != nil {
		panic(err)
	}
	testImage = img
}

func BenchmarkGetPixelsNoRowPitch(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		core.GetPixels(testImage, 0)
	}
}

func BenchmarkGetPixelsSmallRowPitch(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		core.GetPixels(testImage, 4)
	}
}

func BenchmarkGetPixelsMediumRowPitch(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		core.GetPixels(testImage, 200)
	}
}

func BenchmarkGetPixelsBigRowPitch(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		core.GetPixels(testImage, 1000)
	}
}
