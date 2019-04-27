package device

import (
	"fmt"

	vk "github.com/vulkan-go/vulkan"
)

var DefaultVulkanApplicationInfo *vk.ApplicationInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "Koru command line\x00",
	PEngineName:        "https://github.com/koru3d\x00",
}

func NewVulkanDevice(appInfo *vk.ApplicationInfo, window uintptr) (Device, error) {
	v := &Vulkan{}

	var extensions []string
	instanceInfo := &vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        appInfo,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}

	if err := vk.Error(vk.CreateInstance(instanceInfo, nil, &v.instance)); err != nil {
		return nil, err
	} else {
		vk.InitInstance(v.instance)
	}

	if err := v.enumerateDevices(); err != nil {
		return nil, err
	}

	return v, nil
}

type Vulkan struct {
	Device

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

func (v *Vulkan) PhysicalDevices() []PhysicalDeviceInfo {
	pdi := make([]PhysicalDeviceInfo, len(v.availableDevices))
	for i, _ := range pdi {
		pdi[i].Invalid = false
	}

	for i := 0; i < len(v.availableDevices); i++ {
		var numDeviceExtensions uint32
		if err := vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, nil); err != nil {
			pdi[i].Invalid = true
		}
		deviceExt := make([]vk.ExtensionProperties, numDeviceExtensions)
		if err := vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, deviceExt); err != nil {
			pdi[i].Invalid = true
		}
		for _, ext := range deviceExt {
			ext.Deref()
			append(pdi[i].Extensions, vk.ToString(ext.ExtensionName[:])
		}

		var numDeviceLayers uint32
		if err := vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, nil); err != nil {
			pdi[i].Invalid = true
		}
		deviceLayers := make([]vk.LayerProperties, numDeviceLayers)
		if err := vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, deviceLayers); err != nil {
			pdi[i].Invalid = true
		}
		for _, layer := range deviceLayers {
			layer.Deref()
			append(pdi[i].Layers, vk.ToString(layer.LayerName[:]))
		}
		var memoryProperties vk.PhysicalDeviceMemoryProperties
		if err := vk.GetPhysicalDeviceMemoryProperties(v.availableDevices[i], &memoryProperties); err != nil {
			pdi[i].Invalid = true
		}

		memoryProperties.Deref()
		for iMem := 0; iMem < memoryProperties.MemoryHeapCount, iMem++ {
			pdi[i].Memory = pdi[i].Memory + memoryProperties.MemoryHeaps[iMem].DeviceSize
		}
	}
	return pd
}

func (vd *Vulkan) Destroy() {
	if vd == nil {
		return
	}
	vd.availableDevices = nil
	vk.DestroyDevice(vd.device, nil)
	vk.DestroyInstance(vd.instance, nil)
}
