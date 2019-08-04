package model

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/devblok/koru/util/collada"
	glm "github.com/go-gl/mathgl/mgl32"
)

// ImportColladaObject reads given file and converts Collada object to
// engine's internal object
func ImportColladaObject(fileContents []byte) (Object, error) {
	var colladaModel collada.Collada
	if err := xml.Unmarshal(fileContents, &colladaModel); err != nil {
		return nil, err
	}

	mesh := colladaModel.Geometries[0].Mesh
	source, err := findSource(mesh.Source, "positions")
	if err != nil {
		return nil, err
	}

	var vertices []Vertex
	stride := 6
	for idx := 0; idx < len(mesh.Triangles.Index)/stride; idx++ {
		var vert Vertex
		indices := mesh.Triangles.Index[stride*idx : (stride*idx)+stride]
		vert.Pos = glm.Vec3{
			source.Floats.Data[indices[0]],
			source.Floats.Data[indices[1]],
			source.Floats.Data[indices[2]],
			// Other 3 elements is a Vec3 for the vertice's normal
		}
		vert.Color = glm.Vec4{1.0, 1.0, 0.0, 1.0}
		vertices = append(vertices, vert)
	}

	return &ColladaObject{
		vertices: vertices,
	}, nil
}

// ColladaObject is imported from a collada (.dae) file.
// Loaded and held in memory
type ColladaObject struct {
	Object

	mutex    sync.RWMutex
	position glm.Mat4
	rotation glm.Mat4

	vertices []Vertex
}

// SetPosition implements interface
func (co *ColladaObject) SetPosition(pos glm.Mat4) {
	co.mutex.Lock()
	co.position = pos
	co.mutex.Unlock()
}

// Position implements interface
func (co *ColladaObject) Position() glm.Mat4 {
	co.mutex.RLock()
	defer co.mutex.RUnlock()
	return co.position
}

// SetRotation implements interface
func (co *ColladaObject) SetRotation(rot glm.Mat4) {
	co.mutex.Lock()
	co.position = rot
	co.mutex.Unlock()
}

// Rotation implements interface
func (co *ColladaObject) Rotation() glm.Mat4 {
	co.mutex.RLock()
	defer co.mutex.RUnlock()
	return co.rotation
}

// Vertices implements interface
func (co *ColladaObject) Vertices() []Vertex {
	return co.vertices
}

func findSource(sources []collada.Source, dataType string) (collada.Source, error) {
	for _, s := range sources {
		if strings.HasSuffix(s.ID, fmt.Sprintf("-%s", dataType)) {
			return s, nil
		}
	}
	return collada.Source{}, errors.New("source type not found")
}
