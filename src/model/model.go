// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

import (
	"image"
	"unsafe"

	vk "github.com/devblok/vulkan"
	glm "github.com/go-gl/mathgl/mgl32"
)

// Object represents the engine supported model
type Object interface {

	// SetPosition sets the object's current position in space.
	// Has to be thread-safe
	SetPosition(glm.Mat4)

	// Position gets the object's current position in space.
	// Has to be thread-safe
	Position() glm.Mat4

	// SetRotation sets the object's rotation matrix.
	// Has to be thread-safe
	SetRotation(glm.Mat4)

	// Rotation gets the object's rotation matrix.
	// Has to be thread-safe
	Rotation() glm.Mat4

	// Vertices returns the vertices for Renderer use,
	// so it has to match the descriptors exactly
	Vertices() []Vertex

	// Texture returns the raw data of a color texture image
	// for use in the Renderer
	Texture() image.Image
}

// Vertex is a model vertex
type Vertex struct {
	Pos    glm.Vec3
	Normal glm.Vec3
	Color  glm.Vec4
	Tex    glm.Vec2
}

// Texture is a container component for textures.
// Because textures can either be in an image form or raw form,
// this struct accomodates both
type Texture struct {
	Img    image.Image
	Raw    []uint8
	Bounds image.Rectangle
}

// Uniform defines a model-view-projection object
type Uniform struct {
	View       glm.Mat4
	Projection glm.Mat4
}

// VertexBindingDescriptions return Vulkan Vertex descriptors
func VertexBindingDescriptions() []vk.VertexInputBindingDescription {
	return []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    uint32(unsafe.Sizeof(Vertex{})),
		InputRate: vk.VertexInputRateVertex,
	}}
}

// VertexAttributeDescriptions return Vulkan attribute descriptors
func VertexAttributeDescriptions() []vk.VertexInputAttributeDescription {
	return []vk.VertexInputAttributeDescription{
		{
			Binding:  0,
			Location: 0,
			Format:   vk.FormatR32g32b32Sfloat,
			Offset:   uint32(unsafe.Offsetof(Vertex{}.Pos)),
		},
		{
			Binding:  0,
			Location: 1,
			Format:   vk.FormatR32g32b32a32Sfloat,
			Offset:   uint32(unsafe.Offsetof(Vertex{}.Color)),
		},
		{
			Binding:  0,
			Location: 2,
			Format:   vk.FormatR32g32Sfloat,
			Offset:   uint32(unsafe.Offsetof(Vertex{}.Tex)),
		},
	}
}
