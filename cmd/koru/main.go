package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/devblok/koru/core"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	runtime.LockOSThread()
}

// Essential globals
var (
	vkInstance core.Instance
	vkRenderer core.Renderer
	sdlWindow  *sdl.Window
	sdlSurface unsafe.Pointer
)

var configuration = core.Configuration{
	Time: core.TimeConfiguration{
		FramesPerSecond: 60,
	},
	Renderer: core.RendererConfiguration{
		ScreenWidth:   800,
		ScreenHeight:  600,
		SwapchainSize: 3,
		DeviceExtensions: []string{
			"VK_KHR_swapchain",
		},
		ShaderDirectory: "./shaders",
	},
}

func newWindow() *sdl.Window {
	window, err := sdl.CreateWindow("Koru3D",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int32(configuration.Renderer.ScreenWidth),
		int32(configuration.Renderer.ScreenHeight),
		sdl.WINDOW_VULKAN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	return window
}

func handleWindowEvent(event *sdl.WindowEvent) {
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	if err := sdl.VulkanLoadLibrary(""); err != nil {
		panic(err)
	}
	defer sdl.VulkanUnloadLibrary()

	{
		cfg := core.InstanceConfiguration{
			DebugMode:  true,
			Extensions: sdlWindow.VulkanGetInstanceExtensions(),
			Layers:     []string{},
		}

		if vi, err := core.NewVulkanInstance(core.DefaultVulkanApplicationInfo, sdl.VulkanGetVkGetInstanceProcAddr(), cfg); err != nil {
			panic(err)
		} else {
			vkInstance = vi
		}
	}

	sdlWindow = newWindow()
	if srf, err := sdlWindow.VulkanCreateSurface(vkInstance.Instance()); err != nil {
		panic(err)
	} else {
		sdlSurface = srf
		vkInstance.SetSurface(sdlSurface)
	}

	var rendererErr error
	vkRenderer, rendererErr = core.NewVulkanRenderer(vkInstance, configuration.Renderer)
	if rendererErr != nil {
		panic(rendererErr)
	}

	deviceUsed := vkInstance.AvailableDevices()[0]
	if suitable, reason := vkRenderer.DeviceIsSuitable(deviceUsed); !suitable {
		panic(reason)
	}

	if err := vkRenderer.Initialise(); err != nil {
		panic(err)
	}

	time := core.NewTime(configuration.Time)
	exitC := make(chan struct{}, 2)

EventLoop:
	for {
		select {
		case <-exitC:
			log.Println("Event loop exited")
			break EventLoop
		case <-time.FpsTicker().C:
			var event sdl.Event
			for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch et := event.(type) {
				case *sdl.WindowEvent:
					handleWindowEvent(et)
					log.Println("Window event caught")
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
			if err := vkRenderer.Draw(); err != nil {
				log.Println("Draw error: " + err.Error())
			}
		}
	}

	vkRenderer.Destroy()
	vkInstance.Destroy()
}
