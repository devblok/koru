package core

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

const shaderSuffix = ".spv"

// loadShaderFilesFromDirectory get the list of files that are compiled shaders
// it is important that the file name does not contain more than two dots,
// the first is always the name of the shader, second is type, and the third one
// ensured that the shader is compiled (only compiled shaders have an .spv extension).
// All shader files will be loaded.
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

			if len(nodes) != 2 {
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

// SliceUint32 reslices bytes into a uint32, that is used
// to sumbit vulkan shaders for processing
func SliceUint32(data []byte) []uint32 {
	const m = 0x7fffffff
	return (*[m / 4]uint32)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&data)).Data))[:len(data)/4]
}

func safeString(s string) string {
	return fmt.Sprintf("%s\x00", s)
}

func safeStrings(sgs []string) []string {
	safe := []string{}
	for _, s := range sgs {
		safe = append(safe, fmt.Sprintf("%s\x00", s))
	}
	return safe
}

// GetPixels transforms a given image into right arrangement of pixels
// by drawing the decoded image onto a controlled RGBA canvas
func GetPixels(img image.Image, rowPitch int) ([]uint8, error) {
	newImg := image.NewRGBA(img.Bounds())
	if rowPitch <= 4*img.Bounds().Dy() {
		// apply the proposed row pitch only if supported,
		// as we're using only optimal textures.
		newImg.Stride = rowPitch
	}
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)
	return newImg.Pix, nil
}
