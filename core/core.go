package core

import (
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// Destroyable defines a structure which needs to be dismantled
type Destroyable interface {
	// Destroy is used to dismantle the struct in question
	Destroy()
}

// Instance describes a Vulkan instance and supporting methods.
// Once created it is ready to use.
type Instance interface {
	Destroyable

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

	// Instance returns the underlying API instance
	Instance() interface{}

	// Extensions returns available instance extensions
	Extensions() []string
}

// Renderer describes the rendering machinery.
// It's created only with internal values set,
// it needs to be initialised with Initialise() before use.
type Renderer interface {
	Destroyable

	// Initialise sets up the configured rendering pipeline
	Initialise() error

	// DeviceIsSuitable checks if the device given is suitable
	// for the rendering pipeline. If not suitable string contains the reason
	DeviceIsSuitable(vk.PhysicalDevice) (bool, string)
}

// ShaderType represents the type of shader thats loaded
type ShaderType int

// Identifies shader objects with their types
const (
	VertexShaderType ShaderType = iota
	FragmentShaderType
	UnknownShaderType
)

// Shader is an abstraction for shader modules
type Shader interface {
	Destroyable

	// Shader is the internal API shader instance
	ShaderModule() interface{}

	// Type returns the type of shader in question
	Type() ShaderType

	// Name Shader name
	Name() string
}
