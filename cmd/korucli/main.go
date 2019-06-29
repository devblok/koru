package main

import (
	"encoding/json"
	"fmt"

	"github.com/koru3d/koru/core"
)

func main() {
	coreInstance, err := core.NewVulkanInstance(core.DefaultVulkanApplicationInfo, nil, []string{})
	if err != nil {
		panic(err)
	}

	if bytes, err := json.Marshal(coreInstance.PhysicalDevicesInfo()); err == nil {
		fmt.Printf("%s", bytes)
	} else {
		panic(err)
	}

	coreInstance.Destroy()
}
