package device

import vk "github.com/vulkan-go/vulkan"

type PhysicalDeviceInfo struct {
	Id            int
	VendorId      int
	DriverVersion int
	Name          string
	Invalid       bool
	Extensions    []string
	Layers        []string
	Memory        vk.DeviceSize
}

type Device interface {
	PhysicalDevices() []PhysicalDeviceInfo
	Destroy()
}
