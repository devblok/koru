package core

import (
	"unsafe"

	"github.com/devblok/koru/model"
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

	// BuildResources creates a ResourceBuilder that deals with instantiating a new model
	BuildResources() ResourceBuilder

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

// RendererResources is a container for Renderer resources (device memory, textures etc). It is an *instance* of resources.
// This interface should be given to every entity/model that has resources on the renderer's side.
// A RendererResource should be shared between instances that utilize the same set of resources for performance reasons (TODO).
// Destroying the RendererResources should automatically remove it from rendering and destroy ascociated
// memory etc., it is the normal way to do this. It is important it does not remove itself if any other entities
// in the Renderer still have it attached. Destroy should have no effect in that case.
// A Resource is created with a ResourceBuilder.
type RendererResources interface {
	Destroyable

	// Hidden sets if the resouce participates in rendering or not
	Hidden(bool)

	// IsHidden returns the hidden state of the resources
	IsHidden() bool
}

// ResourceBuilder builds RendererResources based on given data
type ResourceBuilder interface {

	// Build constructs and returns the set of resources queried
	Build() (RendererResources, error)

	// WithModel builds resources with a given model
	WithModel(model.Object) ResourceBuilder

	// StartHidden creates the resources, but they will not participate in rendering outright,
	// this behaviour is changed on RendererResources Hidden(bool) member
	StartHidden(bool) ResourceBuilder
}
