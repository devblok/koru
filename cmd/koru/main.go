package main

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/koru3d/koru/device"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	"github.com/xlab/closer"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	var procAddr unsafe.Pointer
	if procAddr := glfw.GetVulkanGetInstanceProcAddress(); procAddr == nil {
		panic(errors.New("glfw proc address was nil"))
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	vkDevice := device.NewVulkanDevice(device.DefaultVulkanApplicationInfo, procAddr)

	defer closer.close()

}
