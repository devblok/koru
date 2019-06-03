package main

import (
	"encoding/json"
	"fmt"

	"github.com/koru3d/koru/device"
)

func main() {
	appDevice, err := device.NewVulkanDevice(device.DefaultVulkanApplicationInfo, nil)
	if err != nil {
		panic(err)
	}

	if bytes, err := json.Marshal(appDevice.PhysicalDevices()); err == nil {
		fmt.Printf("%s", bytes)
	} else {
		panic(err)
	}

	appDevice.Destroy()
}
