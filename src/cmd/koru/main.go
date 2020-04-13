// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/devblok/koru/src/core"
	glm "github.com/go-gl/mathgl/mgl32"
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

	frameCounter int64
)

// Profiling
var (
	cpuProfile   = flag.String("cpuprof", "", "Profile CPU usage to file")
	memProfile   = flag.String("memprof", "", "Profile memory usage into a file")
	traceProfile = flag.String("trace", "", "Trace output for profiling")
	debug        = flag.Bool("vkdbg", false, "Load Vulkan validation layers")
)

var configuration = core.Configuration{
	Time: core.TimeConfiguration{
		FramesPerSecond: 2000,
		EventPollDelay:  50,
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

var constant float32

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

func main() {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *traceProfile != "" {
		f, err := os.Create(*traceProfile)
		if err != nil {
			panic(err)
		}
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

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
			DebugMode:  *debug,
			Extensions: sdlWindow.VulkanGetInstanceExtensions(),
			Layers:     []string{},
		}

		if vi, err := core.NewVulkanInstance(core.DefaultVulkanApplicationInfo, sdl.VulkanGetVkGetInstanceProcAddr(), cfg); err != nil {
			panic(err)
		} else {
			vkInstance = vi
		}
		defer vkInstance.Destroy()
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
	defer vkRenderer.Destroy()

	srh := vkRenderer.ResourceHandle()
	crh := vkRenderer.ResourceHandle()

	timeService := core.NewTime(configuration.Time)

	ctx, cancel := context.WithCancel(context.Background())

	programSync := sync.WaitGroup{}

	/* Frame counter loop */
	programSync.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
	CounterLoop:
		for {
			select {
			case <-ctx.Done():
				break CounterLoop
			default:
				currentCount := atomic.LoadInt64(&frameCounter)
				atomic.StoreInt64(&frameCounter, 0)
				fmt.Printf("\r\033[2KFrame count: %d\tCGO calls: %d", currentCount*5, runtime.NumCgoCall())
				time.Sleep(200 * time.Millisecond)
				// 200 ms * 5 = 1s, therefore we need to mutiply the count
			}
		}
		wg.Done()
	}(ctx, &programSync)

	/* Renderer loop */
	programSync.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
	DrawLoop:
		for {
			select {
			case <-ctx.Done():
				log.Println("Event loop exited")
				break DrawLoop
			case <-timeService.FpsTicker().C:
				if _, ok := <-vkRenderer.ResourceUpdate(srh, core.ResourceInstance{
					ResourceID: "assets/suzanne.dae",
					Position:   glm.Translate3D(0, 0, 0),
					Rotation:   glm.HomogRotate3D(constant, glm.Vec3{0, 0, 1}),
				}); !ok {
					fmt.Printf("Error: not updated resource\n")
				}
				if _, ok := <-vkRenderer.ResourceUpdate(crh, core.ResourceInstance{
					ResourceID: "assets/cube.dae",
					Position:   glm.Translate3D(0, 0, 0),
					Rotation:   glm.HomogRotate3D(constant, glm.Vec3{0, 0, 1}),
				}); !ok {
					fmt.Printf("Error: not updated resource\n")
				}
				constant += 0.005
				if err := vkRenderer.Draw(); err != nil {
					log.Println("Draw error: " + err.Error())
				}
				if err := vkRenderer.Present(); err != nil {
					log.Println("Present error: " + err.Error())
				}
				atomic.AddInt64(&frameCounter, 1)
			}
		}
		wg.Done()
	}(ctx, &programSync)

	/* Event loop */
EventLoop:
	for {
		select {
		case <-ctx.Done():
			break EventLoop
		case <-timeService.EventTicker().C:
			var event sdl.Event
			for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch et := event.(type) {
				case *sdl.KeyboardEvent:
					if et.Keysym.Sym == sdl.K_ESCAPE {
						cancel()
						continue EventLoop
					}
				case *sdl.QuitEvent:
					cancel()
					continue EventLoop
				}
			}
		}
	}

	programSync.Wait()

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			panic(err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
	}
}
