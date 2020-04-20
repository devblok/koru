// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package vkr

import (
	"errors"
	"fmt"
	"unsafe"

	vk "github.com/devblok/vulkan"
)

// Memory defines a usable memory region.
type Memory struct {
	mapped      bool
	len, offset uint
	device      vk.Device
	memory      vk.DeviceMemory
}

// Len returns the length of assigned memory.
func (m *Memory) Len() uint {
	return m.len
}

// Offset returns the start location of assigned memory.
func (m *Memory) Offset() uint {
	return m.offset
}

// Get returns the vulkan memory handle.
func (m *Memory) Get() vk.DeviceMemory {
	return m.memory
}

// Map maps the entire available memory region and
// returns a pointer to the mapped area.
func (m *Memory) Map() unsafe.Pointer {
	var memMapped unsafe.Pointer
	vk.MapMemory(m.device, m.memory, vk.DeviceSize(m.offset), vk.DeviceSize(m.len), 0, &memMapped)
	m.mapped = true
	return memMapped
}

// Unmap removes the memory mapping if it was mapped.
func (m *Memory) Unmap() {
	if m.mapped {
		vk.UnmapMemory(m.device, m.memory)
		m.mapped = false
	}
}

// Release frees memory after unmapping it if previously mapped.
func (m *Memory) Release() {
	m.Unmap()
	vk.FreeMemory(m.device, m.memory, nil)
}

// NewMemoryAllocator creates a new memory allocator. Allocates for the logical device,
// reads memory properties of the physical device to influence allocation.
func NewMemoryAllocator(device vk.Device, phyDevice vk.PhysicalDevice) (*MemoryAllocator, error) {
	var memProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(phyDevice, &memProperties)
	memProperties.Deref()

	return &MemoryAllocator{
		device:        device,
		memProperties: memProperties,
	}, nil
}

// MemoryAllocator is responsible returning usable
// memory for any resources that may need it.
type MemoryAllocator struct {
	device        vk.Device
	memProperties vk.PhysicalDeviceMemoryProperties
}

// Malloc returns a usable memory chunk ready for use.
func (ma *MemoryAllocator) Malloc(req vk.MemoryRequirements, prop vk.MemoryPropertyFlagBits) (Memory, error) {
	memTypeIdx, err := ma.findMemoryType(req.MemoryTypeBits, vk.MemoryPropertyFlags(prop))
	if err != nil {
		return Memory{}, err
	}

	mai := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  req.Size,
		MemoryTypeIndex: memTypeIdx,
	}

	var memory vk.DeviceMemory
	if err := vk.Error(vk.AllocateMemory(ma.device, &mai, nil, &memory)); err != nil {
		return Memory{}, fmt.Errorf("vk.AllocateMemory(): %s", err.Error())
	}

	return Memory{
		offset: 0,
		len:    uint(req.Size),
		device: ma.device,
		memory: memory,
	}, nil
}

func (ma *MemoryAllocator) findMemoryType(filter uint32, prop vk.MemoryPropertyFlags) (uint32, error) {
	for idx := uint32(0); idx < ma.memProperties.MemoryTypeCount; idx++ {
		ma.memProperties.MemoryTypes[idx].Deref()
		if filter&(1<<idx) != 0 && (ma.memProperties.MemoryTypes[idx].PropertyFlags&prop) == prop {
			return idx, nil
		}
	}
	return 0, errors.New("suitable memory type not found")
}
