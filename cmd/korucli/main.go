package main

import (
	"encoding/json"
	"fmt"

	"github.com/koru3d/koru/core"
)

func main() {
	appDevice, err := core.NewVulkanInstance(core.DefaultVulkanApplicationInfo, nil, []string{})
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
