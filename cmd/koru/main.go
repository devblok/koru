package main

import (
	"log"
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
	vkInstance core.Instance
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
	extensions = append(extensions, "VK_KHR_swapchain")
	if vi, err := core.NewVulkanInstance(
		core.DefaultVulkanApplicationInfo,
		sdl.VulkanGetVkGetInstanceProcAddr(),
		extensions); err != nil {
		panic(err)
	} else {
		vkInstance = vi
	}
	defer closer.Close()

	sdlWindow = newWindow()
	if srf, err := sdlWindow.VulkanCreateSurface(vkInstance.Inner()); err != nil {
		panic(err)
	} else {
		sdlSurface = srf
	}

	time := core.NewTime(core.TimeConfiguration{
		FramesPerSecond: 60,
	})
	exitC := make(chan struct{}, 2)

EventLoop:
	for {
		select {
		case <-exitC:
			log.Println("Even loop exited")
			break EventLoop
		case <-time.FpsTicker().C:
			var event sdl.Event
			for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch et := event.(type) {
				case *sdl.KeyboardEvent:
					if et.Keysym.Sym == sdl.K_ESCAPE {
						exitC <- struct{}{}
						continue EventLoop
					}
				case *sdl.QuitEvent:
					exitC <- struct{}{}
					continue EventLoop
				}
			}
		}
	}
}
