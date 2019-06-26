package device

import vk "github.com/vulkan-go/vulkan"

// PhysicalDeviceInfo describes available physical properties of a rendering device
type PhysicalDeviceInfo struct {
	ID            int
	VendorID      int
	DriverVersion int
	Name          string
	Invalid       bool
	Extensions    []string
	Layers        []string
	Memory        vk.DeviceSize
	Features      vk.PhysicalDeviceFeatures
}

// Device describes a non-concrete rendering device
type Device interface {
	PhysicalDevices() []PhysicalDeviceInfo
	Instance() interface{}
	Destroy()
}
