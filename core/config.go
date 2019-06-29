package core

// Configuration defines a global engine configuration setting
type Configuration struct {
	Time     TimeConfiguration
	Renderer RendererConfiguration
}

// TimeConfiguration is used to configure time services
type TimeConfiguration struct {
	// FramesPerSecond caps frames per second that is put out
	// To unlimit, set to 0
	FramesPerSecond int
}

// RendererConfiguration is used to configure the renderer
type RendererConfiguration struct {
	SwapchainSize    uint32
	DeviceExtensions []string

	ScreenWidth  uint32
	ScreenHeight uint32
}
