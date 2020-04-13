// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"

	"github.com/devblok/koru/src/core"
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
