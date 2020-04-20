// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package vkr implements the vulkan renderer.
package vkr

import (
	"fmt"

	"github.com/devblok/koru/src/gfx"
	vk "github.com/devblok/vulkan"
)

// NewBuffer creates, configures, allocates and binds a new buffer.
func NewBuffer(dev vk.Device, size uint, usage vk.BufferUsageFlagBits, mode vk.SharingMode, ma *MemoryAllocator) (Buffer, error) {
	createInfo := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(size),
		Usage:       vk.BufferUsageFlags(usage),
		SharingMode: mode,
	}
	var buffer vk.Buffer
	if err := vk.Error(vk.CreateBuffer(dev, &createInfo, nil, &buffer)); err != nil {
		return Buffer{}, fmt.Errorf("vk.CreateBuffer(): %s", err.Error())
	}

	var req vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(dev, buffer, &req)
	req.Deref()

	memory, err := ma.Malloc(req, vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)
	if err != nil {
		return Buffer{}, err
	}

	vk.BindBufferMemory(dev, buffer, memory.Get(), vk.DeviceSize(memory.Offset()))

	return Buffer{
		device: dev,
		buffer: buffer,
		memory: memory,
	}, nil
}

// Buffer implements a generic vulkan buffer.
type Buffer struct {
	device vk.Device
	buffer vk.Buffer

	memory Memory
}

// Mem returns the Memory that the buffer is based on.
func (b *Buffer) Mem() *Memory {
	return &b.memory
}

// Get returns the vulkan Buffer handle.
func (b *Buffer) Get() vk.Buffer {
	return b.buffer
}

// Release destroys the buffer and memory asociated with it.
func (b *Buffer) Release() {
	vk.DestroyBuffer(b.device, b.buffer, nil)
	b.memory.Release()
}

// NewImage creates a new vulkan image primitive.
func NewImage(dev vk.Device, extent gfx.Extent3D, usage vk.ImageUsageFlagBits, mode vk.SharingMode) (Image, error) {
	createInfo := vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Extent: vk.Extent3D{
			Width:  uint32(extent.Width),
			Height: uint32(extent.Height),
			Depth:  uint32(extent.Depth),
		},
		MipLevels:     1,
		ArrayLayers:   1,
		Format:        vk.FormatR8g8b8a8Unorm,
		Tiling:        vk.ImageTilingLinear,
		InitialLayout: vk.ImageLayoutUndefined,
		Usage:         vk.ImageUsageFlags(usage),
		SharingMode:   mode,
		Samples:       vk.SampleCount1Bit,
	}

	var image vk.Image
	if err := vk.Error(vk.CreateImage(dev, &createInfo, nil, &image)); err != nil {
		return Image{}, fmt.Errorf("vk.CreateImage(): %s", err.Error())
	}

	return Image{
		image: image,
	}, nil
}

// Image implements and abstracts vulkan image primitive.
type Image struct {
	image  vk.Image
	memory Memory
}

// Mem returns the underlying memory of the Image.
func (i *Image) Mem() *Memory {
	return &i.memory
}
