package collada_test

import (
	"encoding/xml"
	"testing"

	"github.com/devblok/koru/util/collada"
)

func TestTrianglesDecode(t *testing.T) {
	data := `
		<triangles material="Material-material" count="12">
		<input semantic="VERTEX" source="#Cube-mesh-vertices" offset="0"/>
		<input semantic="NORMAL" source="#Cube-mesh-normals" offset="1"/>
		<p>0 0 2 0 3 0 7 1 5 1 4 1 4 2 1 2 0 2 5 3 2 3 1 3 2 4 7 4 3 4 0 5 7 5 4 5 0 6 1 6 2 6 7 7 6 7 5 7 4 8 5 8 1 8 5 9 6 9 2 9 2 10 6 10 7 10 0 11 3 11 7 11</p>
		</triangles>
	`
	var triangles collada.Triangles
	err := xml.Unmarshal([]byte(data), &triangles)
	if err != nil {
		t.Fatal(err)
	}

	if triangles.Material != "Material-material" {
		t.Fatalf("incorrect material: %s", triangles.Material)
	}

	if triangles.Count != 12 {
		t.Fatalf("incorrect count: %d", triangles.Count)
	}

	if len(triangles.Inputs) != 2 {
		t.Fatalf("number of inputs incorrect: %d", len(triangles.Inputs))
	}

	if len(triangles.Index) != 12*6 {
		t.Fatalf("number of index elements incorrect: %d", len(triangles.Index))
	}
}

func TestInputDecode(t *testing.T) {
	data := `
	<object>
		<input semantic="VERTEX" source="#Cube-mesh-vertices" offset="0" />
		<input semantic="NORMAL" source="#Cube-mesh-normals" offset="1" />
		<input semantic="TEXTUR" source="#Cube-mesh-textures" offset="2" />
	</object>
	`

	type Object struct {
		XMLNname xml.Name        `xml:"object"`
		Inputs   []collada.Input `xml:"input"`
	}

	var obj Object
	err := xml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if obj.Inputs[0].Offset != 0 || obj.Inputs[0].Semantic != "VERTEX" || obj.Inputs[0].Source != "#Cube-mesh-vertices" {
		t.Logf("expected Offset: %d, got: %d\nexpected Semantic: %s, got %s\nexpected Source: %s, got: %s", 0, obj.Inputs[0].Offset, "VERTEX", obj.Inputs[0].Semantic, "#Cube-mesh-vertices", obj.Inputs[0].Source)
		t.Fail()
	}
	if obj.Inputs[1].Offset != 1 || obj.Inputs[1].Semantic != "NORMAL" || obj.Inputs[1].Source != "#Cube-mesh-normals" {
		t.Logf("expected Offset: %d, got: %d\nexpected Semantic: %s, got %s\nexpected Source: %s, got: %s", 1, obj.Inputs[1].Offset, "NORMAL", obj.Inputs[1].Semantic, "#Cube-mesh-normals", obj.Inputs[1].Source)
		t.Fail()
	}
	if obj.Inputs[2].Offset != 2 || obj.Inputs[2].Semantic != "TEXTUR" || obj.Inputs[2].Source != "#Cube-mesh-textures" {
		t.Logf("expected Offset: %d, got: %d\nexpected Semantic: %s, got %s\nexpected Source: %s, got: %s", 2, obj.Inputs[2].Offset, "TEXTUR", obj.Inputs[2].Semantic, "#Cube-mesh-textures", obj.Inputs[2].Source)
		t.Fail()
	}
}

func TestFloatsDecode(t *testing.T) {
	data := `<float_array id="Cube-mesh-normals-array" count="36">0 0 -1 0 0 1 1 0 -2.38419e-7 0 -1 -4.76837e-7 -1 2.38419e-7 -1.49012e-7 2.68221e-7 1 2.38419e-7 0 0 -1 0 0 1 1 -5.96046e-7 3.27825e-7 -4.76837e-7 -1 0 -1 2.38419e-7 -1.19209e-7 2.08616e-7 1 0</float_array>`

	var floats collada.Floats
	if err := xml.Unmarshal([]byte(data), &floats); err != nil {
		t.Fatal(err)
	}

	if len(floats.Data) != 36 {
		t.Fatalf("bad number of floats, got: %d", len(floats.Data))
	}

	if floats.ID != "Cube-mesh-normals-array" {
		t.Fatalf("bad id, got: %s", floats.ID)
	}
}
