// Package gfx defines rendering related features that renderers must implement.
package gfx

// Releasable defines any memory-occupying item that can be freed.
type Releasable interface {

	// Release releases memory occupied by the implementing structure.
	Release()
}

// Resource describes a rendering resource that can be uniquely identified.
// Can contain multiple resources under itself and combine their operation.
type Resource interface {
	Releasable

	// ID returns a resource id that uniquely identifies it.
	ID() string

	// Ready returns a channel that will be closed when the
	// resource is ready for use.
	Ready() chan<- chan struct{}

	// Sub returns subresources of this resource.
	Sub() []Resource
}

// Loader describes a resource loader mechanism.
type Loader interface {

	// Load tries to find and load the resource
	// asociated with the provided id.
	Load(id string) (Resource, error)
}
