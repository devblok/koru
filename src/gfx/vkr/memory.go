package vkr

import (
	"errors"
	"fmt"

	vk "github.com/devblok/vulkan"
)

// Memory defines a usable memory region.
type Memory struct {
	device vk.Device
	memory vk.DeviceMemory
}

// Get returns the vulkan memory handle.
func (m *Memory) Get() vk.DeviceMemory {
	return m.memory
}

// Release frees memory.
func (m *Memory) Release() {
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
