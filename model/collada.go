package model

import (
	"encoding/xml"
	"fmt"
	"image"
	"strconv"
	"strings"
	"sync"

	glm "github.com/go-gl/mathgl/mgl32"
)

// ImportColladaObject reads given file and converts Collada object to
// engine's internal object
func ImportColladaObject(fileContents []byte, texture image.Image) (Object, error) {
	var colladaModel Collada
	if err := xml.Unmarshal(fileContents, &colladaModel); err != nil {
		return nil, err
	}
	mesh := colladaModel.Geometries[0].Mesh

	// make a map of inputs
	inputs := make(map[uint]Input)
	for _, in := range mesh.Triangles.Inputs {
		inputs[in.Offset] = in
	}

	var vertices []Vertex
	stride := uint(len(mesh.Triangles.Inputs))
	for idx := uint(0); idx < uint(len(mesh.Triangles.Index))/stride; idx++ {
		vertIdx := mesh.Triangles.Index[stride*idx : (stride*idx)+stride]

		var vert Vertex
		for vIdx, v := range vertIdx {
			switch inputs[uint(vIdx)].Semantic {
			case "VERTEX":
				source, err := findSource(mesh.Source, "positions")
				if err != nil {
					return nil, err
				}
				vert.Pos = source.GetVec3(v)
			case "NORMAL":
				source, err := findSource(mesh.Source, "normals")
				if err != nil {
					return nil, err
				}
				vert.Normal = source.GetVec3(v)
			case "TEXCOORD":
				var sourceType string
				if inputs[uint(vIdx)].Set == 0 {
					sourceType = "map"
				} else {
					sourceType = fmt.Sprintf("%s-%d", "map", inputs[uint(vIdx)].Set)
				}
				source, err := findSource(mesh.Source, sourceType)
				if err != nil {
					return nil, err
				}
				vert.Tex = source.GetVec2(v)
			}
		}
		vertices = append(vertices, vert)
	}

	return &ColladaObject{
		vertices: vertices,
		texture:  texture,
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
	texture  image.Image
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

// Texture implements interface
func (co *ColladaObject) Texture() image.Image {
	return co.texture
}

// NormalMap implements interface
func (co *ColladaObject) NormalMap() []byte {
	return []byte{}
}

func findSource(sources []Source, dataType string) (Source, error) {
	for _, s := range sources {
		if strings.HasSuffix(s.ID, fmt.Sprintf("-%s", dataType)) {
			return s, nil
		}
	}
	return Source{}, fmt.Errorf("source type: %s not found", dataType)
}

// Collada is the top-level Collada object
type Collada struct {
	Geometries []Geometry `xml:"library_geometries>geometry"`
	Materials  []Material `xml:"library_materials>material"`
	Effects    []Effect   `xml:"library_effects>effect"`
}

// Geometry represents Collada's geometry
type Geometry struct {
	Mesh Mesh   `xml:"mesh"`
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// Mesh contains all the primitive data
type Mesh struct {
	Source    []Source  `xml:"source"`
	Vertices  Vertices  `xml:"vertices"`
	Triangles Triangles `xml:"triangles"`
}

// Source links to other sources where data is present
type Source struct {
	ID     string `xml:"id,attr"`
	Floats Floats `xml:"float_array"`
	// technique_common define accessing rules, add if needed
}

// GetVec3 returns a set of floats from a given index
// this function assumes array is made in sets of 3 elements
func (s Source) GetVec3(idx int) glm.Vec3 {
	stride := 3
	floats := s.Floats.Data[stride*idx : stride*idx+stride]
	return glm.Vec3{floats[0], floats[1], floats[2]}
}

// GetVec2 returns a set of floats from a given index
// this function assumes array is made in sets of 2 elements
func (s Source) GetVec2(idx int) glm.Vec2 {
	stride := 2
	floats := s.Floats.Data[stride*idx : stride*idx+stride]
	return glm.Vec2{floats[0], floats[1]}
}

// Floats is the array of floats
type Floats struct {
	ID   string
	Data []float32
}

// UnmarshalXML unmarshals the array of floats
func (f *Floats) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			f.ID = attr.Value
		}
	}
	var raw string
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}
	for _, r := range strings.Split(raw, " ") {
		num, err := strconv.ParseFloat(r, 32)
		if err != nil {
			return err
		}
		f.Data = append(f.Data, float32(num))
	}
	return nil
}

// Vertices contains the list of vertices
type Vertices struct {
	ID     string  `xml:"id,attr"`
	Inputs []Input `xml:"input"`
}

// Triangles contain the list of triangles
type Triangles struct {
	Count    int     `xml:"count,attr"`
	Material string  `xml:"material,attr"`
	Inputs   []Input `xml:"input"`
	Index    []int
}

// UnmarshalXML parses the index list
func (t *Triangles) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "count":
			num, err := strconv.Atoi(attr.Value)
			if err != nil {
				return err
			}
			t.Count = num
		case "material":
			t.Material = attr.Value
		}
	}

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch el := token.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "input":
				var input Input
				err := d.DecodeElement(&input, &el)
				if err != nil {
					return err
				}
				t.Inputs = append(t.Inputs, input)
			case "p":
				var (
					ints []int
					raw  string
				)
				if err := d.DecodeElement(&raw, &el); err != nil {
					return err
				}
				for _, r := range strings.Split(raw, " ") {
					num, err := strconv.Atoi(r)
					if err != nil {
						return err
					}
					ints = append(ints, num)
				}
				t.Index = ints
			}
		case xml.EndElement:
			if el == start.End() {
				return nil
			}
		}
	}
}

// Input is Collada'a input type
type Input struct {
	Semantic string `xml:"semantic,attr"`
	Source   string `xml:"source,attr"`
	Offset   uint   `xml:"offset,attr"`
	Set      uint   `xml:"set,attr"`
}

// Material is Collada's material
// located in library_materials
type Material struct {
	ID      string      `xml:"id,attr"`
	Name    string      `xml:"name,attr"`
	Effects []EffectURL `xml:"instance_effect"`
}

// EffectURL contains the location url to the effect
type EffectURL struct {
	URL string `xml:"url,attr"`
}

// Effect is Collada's effect,
// located in library_effects
type Effect struct {
	ID       string     `xml:"id,attr"`
	Emission [4]float32 // The following need custom Unmarshaller
	Diffuse  [4]float32
	Specular float32
}
