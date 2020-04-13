// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

// Configuration defines a global engine configuration setting
type Configuration struct {
	Time     TimeConfiguration
	Instance InstanceConfiguration
	Renderer RendererConfiguration
}

// InstanceConfiguration contains the rendering engine instacne config
type InstanceConfiguration struct {
	DebugMode  bool
	Extensions []string
	Layers     []string
}

// TimeConfiguration is used to configure time services
type TimeConfiguration struct {
	// FramesPerSecond caps frames per second that is put out
	// To unlimit, set to 0
	FramesPerSecond int

	// EventPollDelay configures the event loop with this delay
	// in milliseconds
	EventPollDelay int
}

// RendererConfiguration is used to configure the renderer
type RendererConfiguration struct {
	SwapchainSize    uint32
	DeviceExtensions []string

	ScreenWidth  uint32
	ScreenHeight uint32

	ShaderDirectory string
}
