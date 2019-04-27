package device

type PhysicalDeviceInfo struct {
	Id         int
	Invalid    bool
	Extensions []string
	Layers     []string
	Memory     uint64
}

type Device interface {
	PhysicalDevices() []PhysicalDeviceInfo
	Destroy()
}
