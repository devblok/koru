package main

import (
	"encoding/json"
	"fmt"

	"github.com/koru3d/koru/device"
)

func main() {
	appDevice, err := device.NewVulkanDevice(device.DefaultVulkanApplicationInfo, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", json.Marshal(appDevice.PhysicalDevices()))

	appDevice.Destroy()
}
