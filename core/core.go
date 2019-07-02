package core

import (
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// Instance describes a Vulkan instance and supporting methods.
// Once created it is ready to use.
type Instance interface {
	// PhysicalDevicesInfo returns a struct for each Physical Device
	// along with info about those devices
	PhysicalDevicesInfo() []PhysicalDeviceInfo

	// AvailableDevices returns handles of Physical Devices
	// from the Vulkan API
	AvailableDevices() []vk.PhysicalDevice

	// SetSurface sets the window surface for rendering
	SetSurface(unsafe.Pointer)

	// Surface returns the window surface, if it's not set
	// it should return a valid but empty surface
	Surface() vk.Surface

	// Extensions returns available instance extensions
	Extensions() []string

	// Inner returns the inner handle of the underlying API
	Inner() interface{}

	// Destroy destroys internal members
	Destroy()
}

// Renderer describes the rendering machinery.
// It's created only with internal values set,
// it needs to be initialised with Initialise() before use.
type Renderer interface {
	// Initialise sets up the configured rendering pipeline
	Initialise() error

	// DeviceIsSuitable checks if the device given is suitable
	// for the rendering pipeline. If not suitable string contains the reason
	DeviceIsSuitable(vk.PhysicalDevice) (bool, string)

	// Destroy destroys internal members
	Destroy()
}

// ShaderType represents the type of shader thats loaded
type ShaderType int

// Identifies shader objects with their types
const (
	VertexShaderType ShaderType = iota
	FragmentShaderType
	UnknownShaderType
)
