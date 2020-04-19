// Package vkr implements the vulkan renderer.
package vkr

import (
	"fmt"

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
