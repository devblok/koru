package core

import (
	"fmt"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// DefaultVulkanApplicationInfo application info describes a Vulkan application
var DefaultVulkanApplicationInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "korucli\x00",
	PEngineName:        "Koru3D\x00",
}

// PhysicalDeviceInfo describes available physical properties of a rendering device
type PhysicalDeviceInfo struct {
	ID            int
	VendorID      int
	DriverVersion int
	Name          string
	Invalid       bool
	Extensions    []string
	Layers        []string
	Memory        vk.DeviceSize
	Features      vk.PhysicalDeviceFeatures
}

// Instance describes a Vulkan instance and supporting methods
type Instance interface {
	PhysicalDevices() []PhysicalDeviceInfo
	Inner() interface{}
	Destroy()
}

// NewVulkanInstance creates a Vulkan instance
func NewVulkanInstance(appInfo *vk.ApplicationInfo, window unsafe.Pointer, extensions []string) (Instance, error) {
	if window == nil {
		if err := vk.SetDefaultGetInstanceProcAddr(); err != nil {
			return nil, err
		}
	} else {
		vk.SetGetInstanceProcAddr(window)
	}

	if err := vk.Init(); err != nil {
		return nil, err
	}

	v := &Vulkan{}

	instanceInfo := vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}

	if err := vk.Error(vk.CreateInstance(&instanceInfo, nil, &v.instance)); err != nil {
		return nil, err
	}
	vk.InitInstance(v.instance)

	if err := v.enumerateDevices(); err != nil {
		return nil, err
	}

	return v, nil
}

// Vulkan describes a Vulkan API Instance
type Vulkan struct {
	Instance

	availableDevices []vk.PhysicalDevice

	instance vk.Instance
	surface  vk.Surface
	device   vk.Device
}

func (v *Vulkan) enumerateDevices() error {
	var deviceCount uint32
	if err := vk.Error(vk.EnumeratePhysicalDevices(v.instance, &deviceCount, nil)); err != nil {
		return fmt.Errorf("vulkan physical device enumeration failed: %s", err)
	}
	v.availableDevices = make([]vk.PhysicalDevice, deviceCount)
	if err := vk.Error(vk.EnumeratePhysicalDevices(v.instance, &deviceCount, v.availableDevices)); err != nil {
		return fmt.Errorf("vulkan physical device enumeration failed: %s", err)
	}
	return nil
}

// PhysicalDevices implements interface
func (v *Vulkan) PhysicalDevices() []PhysicalDeviceInfo {
	pdi := make([]PhysicalDeviceInfo, len(v.availableDevices))
	for i := range pdi {
		pdi[i].Invalid = false
	}

	for i := 0; i < len(v.availableDevices); i++ {
		// Get device features
		vk.GetPhysicalDeviceFeatures(v.availableDevices[i], &pdi[i].Features)

		// Get extension info
		var numDeviceExtensions uint32
		if err := vk.Error(vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, nil)); err != nil {
			pdi[i].Invalid = true
		}
		deviceExt := make([]vk.ExtensionProperties, numDeviceExtensions)
		if err := vk.Error(vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, deviceExt)); err != nil {
			pdi[i].Invalid = true
		}
		for _, ext := range deviceExt {
			ext.Deref()
			pdi[i].Extensions = append(pdi[i].Extensions, vk.ToString(ext.ExtensionName[:]))
		}

		// Get layers info
		var numDeviceLayers uint32
		if err := vk.Error(vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, nil)); err != nil {
			pdi[i].Invalid = true
		}
		deviceLayers := make([]vk.LayerProperties, numDeviceLayers)
		if err := vk.Error(vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, deviceLayers)); err != nil {
			pdi[i].Invalid = true
		}
		for _, layer := range deviceLayers {
			layer.Deref()
			pdi[i].Layers = append(pdi[i].Layers, vk.ToString(layer.LayerName[:]))
		}

		// Get memory info
		var memoryProperties vk.PhysicalDeviceMemoryProperties
		vk.GetPhysicalDeviceMemoryProperties(v.availableDevices[i], &memoryProperties)
		memoryProperties.Deref()
		for iMem := (uint32)(0); iMem < memoryProperties.MemoryHeapCount; iMem++ {
			pdi[i].Memory = pdi[i].Memory + memoryProperties.MemoryHeaps[iMem].Size
		}

		// Get general device info
		var physicalDeviceProperties vk.PhysicalDeviceProperties
		vk.GetPhysicalDeviceProperties(v.availableDevices[i], &physicalDeviceProperties)
		physicalDeviceProperties.Deref()
		pdi[i].ID = (int)(physicalDeviceProperties.DeviceID)
		pdi[i].VendorID = (int)(physicalDeviceProperties.VendorID)
		pdi[i].Name = vk.ToString(physicalDeviceProperties.DeviceName[:])
		pdi[i].DriverVersion = (int)(physicalDeviceProperties.DriverVersion)
	}
	return pdi
}

// Inner implements interface
func (v *Vulkan) Inner() interface{} {
	return v.instance
}

// Destroy implements interface
func (v *Vulkan) Destroy() {
	if v == nil {
		return
	}
	v.availableDevices = nil
	vk.DestroyDevice(v.device, nil)
	vk.DestroyInstance(v.instance, nil)
}
