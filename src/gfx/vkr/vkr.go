// Package vkr implements the vulkan renderer.
package vkr

import "github.com/devblok/koru/src/gfx"

// Image implements vulkan image type, use for textures, etc.
type Image struct {
	id string

	ready chan struct{}
}

// ID returns Resource ID.
func (i *Image) ID() string {
	return i.id
}

// Ready returns a channel that is closed when the image
// is in memory and ready for use.
func (i *Image) Ready() <-chan struct{} {
	return i.ready
}

// Sub returns all subresources of this texture.
// Textures cannot have any subresources, so this is always nil.
func (Image) Sub() []gfx.Resource {
	return nil
}

// Release release resources related to the Image.
func (i *Image) Release() {

}

// Mesh implements a vertex collection and related things
// for the vulkan renderer.
type Mesh struct {
	id string

	ready chan struct{}
}

// ID returns Resource ID.
func (m *Mesh) ID() string {
	return m.id
}

// Ready returns a channel that is closed when the image
// is in memory and ready for use.
func (m *Mesh) Ready() <-chan struct{} {
	return m.ready
}

// Sub returns all subresources of this texture.
// Textures cannot have any subresources, so this is always nil.
func (Mesh) Sub() []gfx.Resource {
	return nil
}
