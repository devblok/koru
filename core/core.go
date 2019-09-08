// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"unsafe"

	glm "github.com/go-gl/mathgl/mgl32"
	vk "github.com/devblok/vulkan"
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

// PhysicalDeviceInfo describes available physical properties of a rendering device
type PhysicalDeviceInfo struct {
	ID            int
	VendorID      int
	DriverVersion int
	Name          string
	Invalid       bool
	Extensions    []string
	Layers        []string
	Memory        uint
}

// Renderer describes the rendering machinery.
// It's created only with internal values set,
// it needs to be initialised with Initialise() before use.
type Renderer interface {
	Destroyable

	// Initialise sets up the configured rendering pipeline
	Initialise() error

	// ResourceHandle requests for a unique handle for use with the renderer.
	// Every entity that wishes to be rendered needs to get a unique handle.
	ResourceHandle() ResourceHandle

	// Update updates the current rendering queue at given handle
	ResourceUpdate(ResourceHandle, ResourceInstance) <-chan struct{}

	// ResourceDelete removes the resource from rendering queue
	ResourceDelete(ResourceHandle)

	// Draw draws the frame
	Draw() error

	// Present submits current frame for display
	Present() error

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

// Shader is an abstraction for shader modules.
// It is safe to destroy after the rendering pipeline is created.
type Shader interface {
	Destroyable

	// Shader is the internal API shader instance
	ShaderModule() interface{}

	// Type returns the type of shader in question
	Type() ShaderType

	// Name Shader name
	Name() string
}

// ResourceInstance represents an instance in the renderer.
// Contains all the data, should be updated all at once.
type ResourceInstance struct {

	// ResourceID tells the renderer which resource to load from packages
	ResourceID string

	// Position matrix for the Resource
	Position glm.Mat4

	// Rotation matrix for the Resource
	Rotation glm.Mat4
}

// ResourceHandle identifies the resource instance in the renderer
type ResourceHandle uint32
