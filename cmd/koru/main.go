package main

import (
	"runtime"
	"unsafe"

	"github.com/koru3d/koru/core"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/xlab/closer"
)

func init() {
	runtime.LockOSThread()
}

func newWindow() *sdl.Window {
	window, err := sdl.CreateWindow("Koru3D", sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_VULKAN)
	if err != nil {
		panic(err)
	}
	return window
}

var (
	vkDevice   core.Device
	sdlWindow  *sdl.Window
	sdlSurface unsafe.Pointer
)

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	if err := sdl.VulkanLoadLibrary(""); err != nil {
		panic(err)
	}
	defer sdl.VulkanUnloadLibrary()

	extensions := sdlWindow.VulkanGetInstanceExtensions()
	if vkd, err := core.NewVulkanDevice(
		core.DefaultVulkanApplicationInfo,
		sdl.VulkanGetVkGetInstanceProcAddr(),
		extensions); err != nil {
		panic(err)
	} else {
		vkDevice = vkd
	}
	defer closer.Close()

	sdlWindow = newWindow()
	if srf, err := sdlWindow.VulkanCreateSurface(vkDevice.Instance()); err != nil {
		panic(err)
	} else {
		sdlSurface = srf
	}

}
