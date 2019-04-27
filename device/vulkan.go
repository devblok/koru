package device

import vk "github.com/vulkan-go/vulkan"

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
	return nil
}

func (vd *Vulkan) Destroy() {
	if vd == nil {
		return
	}
	vd.availableDevices = nil
	vk.DestroyDevice(vd.device, nil)
	vk.DestroyInstance(vd.instance, nil)
}
