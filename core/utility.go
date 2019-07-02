package core

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

const shaderSuffix = ".spv"

func loadShaderFilesFromDirectory(dir string) ([]string, []ShaderType, error) {
	var (
		shaders     []string
		shaderTypes []ShaderType
	)
	if err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(f.Name(), shaderSuffix) {
			shader := strings.TrimSuffix(f.Name(), shaderSuffix)
			nodes := strings.Split(shader, ".")

			if len(nodes) < 2 {
				return nil
			}

			suffix := nodes[len(nodes)-1]
			switch suffix {
			case "frag":
				shaderTypes = append(shaderTypes, FragmentShaderType)
				shaders = append(shaders, path)
			case "vert":
				shaderTypes = append(shaderTypes, VertexShaderType)
				shaders = append(shaders, path)
			default:
				return nil
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return shaders, shaderTypes, nil
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func sliceUint32(data []byte) []uint32 {
	const m = 0x7fffffff
	return (*[m / 4]uint32)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&data)).Data))[:len(data)/4]
}
