package collada

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// Collada is the top-level Collada object
type Collada struct {
	Geometries []Geometry `xml:"library_geometries>geometry"`
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
}
