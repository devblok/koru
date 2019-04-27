package device

type PhysicalDeviceInfo struct {
	Id int
}

type Device interface {
	PhysicalDevices() []PhysicalDeviceInfo
	Destroy()
}
