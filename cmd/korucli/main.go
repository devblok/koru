package main

import (
	"encoding/json"
	"fmt"

	"github.com/koru3d/koru/core"
)

func main() {
	cfg := core.InstanceConfiguration{
		DebugMode:  true,
		Extensions: []string{},
		Layers:     []string{},
	}

	coreInstance, err := core.NewVulkanInstance(core.DefaultVulkanApplicationInfo, nil, cfg)
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
