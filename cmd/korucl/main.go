package main

import "github.com/koru3d/koru/device"

func main() {
	appDevice, err := device.NewVulkanDevice(device.DefaultVulkanApplicationInfo, 0)
	if err != nil {
		panic(err)
	}

	appDevice.Destroy()
}
