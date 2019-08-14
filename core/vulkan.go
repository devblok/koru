package core

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/devblok/koru/model"
	glm "github.com/go-gl/mathgl/mgl32"
	vk "github.com/vulkan-go/vulkan"
)

// DefaultVulkanApplicationInfo application info describes a Vulkan application
var DefaultVulkanApplicationInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   safeString("Koru3D"),
	PEngineName:        safeString("Koru3D"),
}

// NewVulkanInstance creates a Vulkan instance
func NewVulkanInstance(appInfo *vk.ApplicationInfo, window unsafe.Pointer, cfg InstanceConfiguration) (Instance, error) {
	if cfg.DebugMode {
		cfg.Layers = append(cfg.Layers, "VK_LAYER_LUNARG_standard_validation")
		cfg.Extensions = append(cfg.Extensions, "VK_EXT_debug_report")
	}

	if window == nil {
		if err := vk.SetDefaultGetInstanceProcAddr(); err != nil {
			return nil, errors.New("vk.InstanceProcAddr(): " + err.Error())
		}
	} else {
		vk.SetGetInstanceProcAddr(window)
	}

	if err := vk.Init(); err != nil {
		return nil, errors.New("vk.Init(): " + err.Error())
	}

	/* Create instance */
	instanceInfo := vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        appInfo,
		EnabledExtensionCount:   uint32(len(cfg.Extensions)),
		PpEnabledExtensionNames: safeStrings(cfg.Extensions),
		EnabledLayerCount:       uint32(len(cfg.Layers)),
		PpEnabledLayerNames:     safeStrings(cfg.Layers),
	}

	var instance vk.Instance
	if err := vk.Error(vk.CreateInstance(&instanceInfo, nil, &instance)); err != nil {
		return nil, errors.New("vk.CreateInstance(): " + err.Error())
	}
	vk.InitInstance(instance)

	/* Enumerate devices */
	physicalDevices, err := enumerateDevices(instance)
	if err != nil {
		return nil, errors.New("core.enumerateDevices(): " + err.Error())
	}

	return &VulkanInstance{
		configuration:    cfg,
		instance:         instance,
		availableDevices: physicalDevices,
	}, nil
}

// VulkanInstance describes a Vulkan API Instance
type VulkanInstance struct {
	Destroyable

	configuration InstanceConfiguration

	availableDevices []vk.PhysicalDevice
	surface          vk.Surface
	instance         vk.Instance
}

func enumerateDevices(instance vk.Instance) ([]vk.PhysicalDevice, error) {
	var deviceCount uint32
	if err := vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)); err != nil {
		return nil, fmt.Errorf("vulkan physical device enumeration failed: %s", err)
	}
	availableDevices := make([]vk.PhysicalDevice, deviceCount)
	if err := vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, availableDevices)); err != nil {
		return nil, fmt.Errorf("vulkan physical device enumeration failed: %s", err)
	}
	return availableDevices, nil
}

// PhysicalDevicesInfo implements interface
func (v VulkanInstance) PhysicalDevicesInfo() []PhysicalDeviceInfo {
	pdi := make([]PhysicalDeviceInfo, len(v.availableDevices))
	for i := 0; i < len(v.availableDevices); i++ {
		// Get extension info
		var numDeviceExtensions uint32
		if err := vk.Error(vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, nil)); err != nil {
			pdi[i].Invalid = true
		}
		deviceExt := make([]vk.ExtensionProperties, numDeviceExtensions)
		if err := vk.Error(vk.EnumerateDeviceExtensionProperties(v.availableDevices[i], "", &numDeviceExtensions, deviceExt)); err != nil {
			pdi[i].Invalid = true
		}
		for _, ext := range deviceExt {
			ext.Deref()
			pdi[i].Extensions = append(pdi[i].Extensions, vk.ToString(ext.ExtensionName[:]))
		}

		// Get layers info
		var numDeviceLayers uint32
		if err := vk.Error(vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, nil)); err != nil {
			pdi[i].Invalid = true
		}
		deviceLayers := make([]vk.LayerProperties, numDeviceLayers)
		if err := vk.Error(vk.EnumerateDeviceLayerProperties(v.availableDevices[i], &numDeviceLayers, deviceLayers)); err != nil {
			pdi[i].Invalid = true
		}
		for _, layer := range deviceLayers {
			layer.Deref()
			pdi[i].Layers = append(pdi[i].Layers, vk.ToString(layer.LayerName[:]))
		}

		// Get memory info
		var memoryProperties vk.PhysicalDeviceMemoryProperties
		vk.GetPhysicalDeviceMemoryProperties(v.availableDevices[i], &memoryProperties)
		memoryProperties.Deref()
		for iMem := (uint32)(0); iMem < memoryProperties.MemoryHeapCount; iMem++ {
			memoryProperties.MemoryHeaps[iMem].Deref()
			pdi[i].Memory = pdi[i].Memory + uint(memoryProperties.MemoryHeaps[iMem].Size)
		}

		// Get general device info
		var physicalDeviceProperties vk.PhysicalDeviceProperties
		vk.GetPhysicalDeviceProperties(v.availableDevices[i], &physicalDeviceProperties)
		physicalDeviceProperties.Deref()
		pdi[i].ID = (int)(physicalDeviceProperties.DeviceID)
		pdi[i].VendorID = (int)(physicalDeviceProperties.VendorID)
		pdi[i].Name = vk.ToString(physicalDeviceProperties.DeviceName[:])
		pdi[i].DriverVersion = (int)(physicalDeviceProperties.DriverVersion)
	}
	return pdi
}

// SetSurface implements interface
func (v *VulkanInstance) SetSurface(pSurface unsafe.Pointer) {
	v.surface = vk.SurfaceFromPointer(uintptr(pSurface))
}

// Surface implements interface
func (v VulkanInstance) Surface() vk.Surface {
	if v.surface == nil {
		return vk.NullSurface
	}
	return v.surface
}

// Instance returns internal vk.Instance
func (v *VulkanInstance) Instance() interface{} {
	return v.instance
}

// Extensions implements interface
func (v VulkanInstance) Extensions() []string {
	return v.configuration.Extensions
}

// AvailableDevices implements interface
func (v VulkanInstance) AvailableDevices() []vk.PhysicalDevice {
	return v.availableDevices
}

// Destroy implements interface
func (v VulkanInstance) Destroy() {
	v.availableDevices = nil
	vk.DestroyInstance(v.instance, nil)
}

// NewVulkanRenderer creates a not yet initialised Vulkan API renderer
func NewVulkanRenderer(instance Instance, cfg RendererConfiguration) (Renderer, error) {
	data, err := ioutil.ReadFile("assets/cube.dae")
	if err != nil {
		return nil, err
	}

	textureFile, err := os.Open("assets/rust.jpeg")
	if err != nil {
		return nil, fmt.Errorf("texture file open failed: %s", err.Error())
	}

	img, err := jpeg.Decode(textureFile)
	if err != nil {
		return nil, fmt.Errorf("jpeg decode failed: %s", err.Error())
	}
	textureFile.Close()

	obj, err := model.ImportColladaObject(data, img)
	if err != nil {
		return nil, fmt.Errorf("collada import failed: %s", err.Error())
	}

	return &VulkanRenderer{
		configuration:        cfg,
		currentSurfaceHeight: cfg.ScreenHeight,
		currentSurfaceWidth:  cfg.ScreenWidth,
		surface:              instance.Surface(),
		physicalDevice:       instance.AvailableDevices()[0],
		vertices:             obj.Vertices(),
		texture:              obj.Texture(),
	}, nil
}

// VulkanRenderer is a Vulkan API renderer
type VulkanRenderer struct {
	Destroyable
	Renderer

	configuration RendererConfiguration

	surface              vk.Surface
	shaders              []Shader
	currentSurfaceHeight uint32
	currentSurfaceWidth  uint32

	swapchain            vk.Swapchain
	swapchainImages      []vk.Image
	swapchainImageViews  []vk.ImageView
	swapchainAttachments []vk.AttachmentDescription
	framebuffers         []vk.Framebuffer

	logicalDevice  vk.Device
	physicalDevice vk.PhysicalDevice
	deviceQueue    vk.Queue

	imageFormat     vk.Format
	imageColorspace vk.ColorSpace

	viewport vk.Viewport
	scissor  vk.Rect2D

	pipelineLayout vk.PipelineLayout
	pipeline       vk.Pipeline
	pipelineCache  vk.PipelineCache

	descriptorPool       vk.DescriptorPool
	descriptorSetLayouts []vk.DescriptorSetLayout
	descriptorSets       []vk.DescriptorSet
	renderPass           vk.RenderPass

	depthImage       vk.Image
	depthImageView   vk.ImageView
	depthImageFormat vk.Format
	depthImageMemory vk.DeviceMemory

	commandPool    vk.CommandPool
	commandBuffers []vk.CommandBuffer

	imageFence              vk.Fence
	renderFinishedSemphore  vk.Semaphore
	imageAvailableSemaphore vk.Semaphore
	imageIndex              uint32

	currentQueueIndex  uint32
	graphicsQueueIndex uint32

	vertices             []model.Vertex
	vertexBuffer         vk.Buffer
	vertexMemory         vk.DeviceMemory
	uniformBuffers       []vk.Buffer
	uniformBuffersMemory []vk.DeviceMemory

	texture            image.Image
	textureBuffer      vk.Buffer
	textureMemory      vk.DeviceMemory
	textureImage       vk.Image
	textureImageMemory vk.DeviceMemory
	textureImageView   vk.ImageView

	textureSampler vk.Sampler
}

// Initialise implements interface
func (v *VulkanRenderer) Initialise() error {
	// TODO: Make extension name escaping bearable
	requiredExtensions := []string{
		vk.KhrSwapchainExtensionName,
	}

	{
		var (
			queueFamilyCount uint32
			queueFamilies    []vk.QueueFamilyProperties
		)
		vk.GetPhysicalDeviceQueueFamilyProperties(v.physicalDevice, &queueFamilyCount, nil)
		queueFamilies = make([]vk.QueueFamilyProperties, queueFamilyCount)
		vk.GetPhysicalDeviceQueueFamilyProperties(v.physicalDevice, &queueFamilyCount, queueFamilies)

		if queueFamilyCount == 0 {
			return errors.New("vk.GetPhysicalDeviceQueueFamilyProperties(): no queuefamilies on GPU")
		}

		/* Find a suitable queue family for the target Vulkan mode */
		var graphicsFound bool
		var presentFound bool
		var separateQueue bool
		for i := uint32(0); i < queueFamilyCount; i++ {
			var (
				required        vk.QueueFlags
				supportsPresent vk.Bool32
				needsPresent    bool
			)
			if graphicsFound {
				// looking for separate present queue
				separateQueue = true
				vk.GetPhysicalDeviceSurfaceSupport(v.physicalDevice, i, v.surface, &supportsPresent)
				if supportsPresent.B() {
					v.currentQueueIndex = i
					presentFound = true
					break
				}
			}

			required |= vk.QueueFlags(vk.QueueGraphicsBit)
			vk.GetPhysicalDeviceSurfaceSupport(v.physicalDevice, i, v.surface, &supportsPresent)
			queueFamilies[i].Deref()
			if queueFamilies[i].QueueFlags&required != 0 {
				if !needsPresent || (needsPresent && supportsPresent.B()) {
					v.graphicsQueueIndex = i
					graphicsFound = true
					break
				} else if needsPresent {
					v.graphicsQueueIndex = i
					graphicsFound = true
					// need present, but this one doesn't support
					// continue lookup
				}
			}
		}
		if separateQueue && !presentFound {
			return errors.New("vulkan error: could not found separate queue with present capabilities")
		}
		if !graphicsFound {
			return errors.New("vulkan error: could not find a suitable queue family for the target Vulkan mode")
		}
	}

	/* Logical Device setup */
	queueInfos := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: 0,
		QueueCount:       1,
		PQueuePriorities: []float32{1, 0},
	}}

	var vkDevice vk.Device
	dci := vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueInfos)),
		PQueueCreateInfos:       queueInfos,
		EnabledExtensionCount:   uint32(len(requiredExtensions)),
		PpEnabledExtensionNames: safeStrings(requiredExtensions),
	}
	if err := vk.Error(vk.CreateDevice(v.physicalDevice, &dci, nil, &vkDevice)); err != nil {
		return errors.New("vk.CreateDevice(): " + err.Error())
	}

	var deviceQueue vk.Queue
	vk.GetDeviceQueue(vkDevice, v.graphicsQueueIndex, 0, &deviceQueue)

	v.deviceQueue = deviceQueue
	v.logicalDevice = vkDevice

	/* ImageFormat */
	var (
		surfaceFormatCount uint32
		surfaceFormats     []vk.SurfaceFormat
	)

	if err := vk.Error(vk.GetPhysicalDeviceSurfaceFormats(v.physicalDevice, v.surface, &surfaceFormatCount, nil)); err != nil {
		return errors.New("vk.GetPhysicalDeviceSurfaceFormats(): " + err.Error())
	}

	surfaceFormats = make([]vk.SurfaceFormat, surfaceFormatCount)
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceFormats(v.physicalDevice, v.surface, &surfaceFormatCount, surfaceFormats)); err != nil {
		return errors.New("vk.GetPhysicalDeviceSurfaceFormats(): " + err.Error())
	}

	surfaceFormats[0].Deref()

	{
		var supported vk.Bool32
		if err := vk.Error(vk.GetPhysicalDeviceSurfaceSupport(v.physicalDevice, 0, v.surface, &supported)); err != nil {
			return errors.New("vk.GetPhysicalDeviceSurfaceSupport(): " + err.Error())
		}

		if !supported.B() {
			return fmt.Errorf("vk.GetPhysicalDeviceSurfaceSupport(): surface is not supported")
		}
	}
	v.imageFormat = surfaceFormats[0].Format
	v.imageColorspace = surfaceFormats[0].ColorSpace

	/* Swapchain setup */
	if err := v.createSwapchain(nil); err != nil {
		return err
	}

	/* Viewport and scissors creation */
	v.createViewport()

	// TODO: Depth and stencil testing VkPipelineDepthStencilStateCreateInfo
	// TODO: When making dynamic state changes refer to  VkPipelineDynamicStateCreateInfo
	// Dynamic state in vulkan-tutorial.com

	/* Depth image */
	if err := v.prepareDepthImage(); err != nil {
		return err
	}

	// /* Uniform buffers */
	// if err := v.prepareUniformBuffers(); err != nil {
	// 	return err
	// }

	/* Pipeline Layout */
	if err := v.createPipelineLayout(); err != nil {
		return err
	}

	/* Render pass */
	if err := v.createRenderPass(); err != nil {
		return err
	}

	/* Shaders */
	if err := v.loadShaders(); err != nil {
		return err
	}

	/* Pipeline cache */
	if err := v.createPipelineCache(); err != nil {
		return err
	}

	/* Pipeline */
	if err := v.createPipeline(); err != nil {
		return err
	}

	if err := v.createImageViews(); err != nil {
		return err
	}

	if err := v.createFramebuffers(); err != nil {
		return err
	}

	if err := v.createCommandPool(); err != nil {
		return err
	}

	if err := v.createTextureImage(); err != nil {
		return err
	}

	if err := v.createTextureImageView(); err != nil {
		return err
	}

	if err := v.createTextureSampler(); err != nil {
		return err
	}

	if err := v.createVertexBuffers(); err != nil {
		return err
	}

	if err := v.createUniformBuffers(); err != nil {
		return err
	}

	if err := v.prepareDescriptorPool(); err != nil {
		return err
	}

	if err := v.createDescriptorSets(); err != nil {
		return err
	}

	if err := v.allocateCommandBuffers(); err != nil {
		return err
	}

	if err := v.createSynchronization(); err != nil {
		return err
	}

	for idx := range v.swapchainImages {
		v.updateUniformBuffers(uint32(idx))
	}

	/* Fill in command buffers */
	if err := v.buildCommandBuffers(); err != nil {
		return err
	}

	return nil
}

func (v *VulkanRenderer) createTextureSampler() error {
	sci := vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               vk.FilterLinear,
		MinFilter:               vk.FilterLinear,
		AddressModeU:            vk.SamplerAddressModeRepeat,
		AddressModeV:            vk.SamplerAddressModeRepeat,
		AddressModeW:            vk.SamplerAddressModeRepeat,
		AnisotropyEnable:        vk.True,
		MaxAnisotropy:           16,
		BorderColor:             vk.BorderColorFloatOpaqueBlack,
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		CompareOp:               vk.CompareOpAlways,
		MipmapMode:              vk.SamplerMipmapModeLinear,
		MipLodBias:              0,
		MinLod:                  0,
		MaxLod:                  0,
	}

	var textureSampler vk.Sampler
	if err := vk.Error(vk.CreateSampler(v.logicalDevice, &sci, nil, &textureSampler)); err != nil {
		return fmt.Errorf("vk.CreateSampler(): %s", err.Error())
	}
	v.textureSampler = textureSampler

	return nil
}

func (v *VulkanRenderer) createTextureImage() error {
	bounds := v.texture.Bounds()
	bufSize := bounds.Max.X * bounds.Max.Y * 4

	var (
		textureBuffer vk.Buffer
		textureMemory vk.DeviceMemory
	)

	bci := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(bufSize),
		Usage:       vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		SharingMode: vk.SharingModeExclusive,
	}
	if err := vk.Error(vk.CreateBuffer(v.logicalDevice, &bci, nil, &textureBuffer)); err != nil {
		return fmt.Errorf("vk.CreateBuffer(): %s", err.Error())
	}

	var memoryRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(v.logicalDevice, textureBuffer, &memoryRequirements)
	memoryRequirements.Deref()
	memTypeIdx, err := findMemoryType(v.physicalDevice, memoryRequirements.MemoryTypeBits, vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	mai := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memoryRequirements.Size,
		MemoryTypeIndex: memTypeIdx,
	}

	if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &mai, nil, &textureMemory)); err != nil {
		return fmt.Errorf("vk.AllocateMemory(): %s", err.Error())
	}

	vk.BindBufferMemory(v.logicalDevice, textureBuffer, textureMemory, 0)

	v.textureBuffer = textureBuffer
	v.textureMemory = textureMemory

	pixels, err := getPixels(v.texture)
	if err != nil {
		return err
	}

	var mappedMemory unsafe.Pointer
	vk.MapMemory(v.logicalDevice, textureMemory, 0, vk.DeviceSize(bufSize), 0, &mappedMemory)
	castMappedMemory := *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: uintptr(mappedMemory),
		Cap:  bufSize,
		Len:  bufSize,
	}))
	copy(castMappedMemory, pixels[:])
	vk.UnmapMemory(v.logicalDevice, textureMemory)

	var (
		textureImage       vk.Image
		textureImageMemory vk.DeviceMemory
	)

	ici := vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Extent: vk.Extent3D{
			Width:  uint32(bounds.Max.X),
			Height: uint32(bounds.Max.Y),
			Depth:  1,
		},
		MipLevels:     1,
		ArrayLayers:   1,
		Format:        vk.FormatR8g8b8a8Snorm,
		Tiling:        vk.ImageTilingOptimal,
		InitialLayout: vk.ImageLayoutUndefined,
		Usage:         vk.ImageUsageFlags(vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit),
		SharingMode:   vk.SharingModeExclusive,
		Samples:       vk.SampleCount1Bit,
	}

	if err := vk.Error(vk.CreateImage(v.logicalDevice, &ici, nil, &textureImage)); err != nil {
		return fmt.Errorf("vk.CreateImage(): %s", err.Error())
	}

	var memRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(v.logicalDevice, textureImage, &memRequirements)
	memRequirements.Deref()

	memIdx, err := findMemoryType(v.physicalDevice, memRequirements.MemoryTypeBits, vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit))
	if err != nil {
		return err
	}

	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memRequirements.Size,
		MemoryTypeIndex: memIdx,
	}

	if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &allocInfo, nil, &textureImageMemory)); err != nil {
		return fmt.Errorf("vk.AllocateMemory(): %s", err.Error())
	}

	vk.BindImageMemory(v.logicalDevice, textureImage, textureImageMemory, 0)

	v.textureImage = textureImage
	v.textureImageMemory = textureImageMemory

	if err := v.transitionLayout(textureImage, vk.FormatR8g8b8a8Unorm, vk.ImageLayoutUndefined, vk.ImageLayoutTransferDstOptimal); err != nil {
		return err
	}

	if err := v.copyBufferToImage(textureBuffer, textureImage, uint32(bounds.Max.X), uint32(bounds.Max.Y)); err != nil {
		return err
	}

	if err := v.transitionLayout(textureImage, vk.FormatR8g8b8a8Unorm, vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal); err != nil {
		return err
	}

	return nil
}

func (v *VulkanRenderer) createTextureImageView() error {
	ivci := vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    v.textureImage,
		ViewType: vk.ImageViewType2d,
		Format:   vk.FormatR8g8b8a8Unorm,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var textureImageView vk.ImageView
	if err := vk.Error(vk.CreateImageView(v.logicalDevice, &ivci, nil, &textureImageView)); err != nil {
		return fmt.Errorf("vk.CreateImageView(): %s", err.Error())
	}
	v.textureImageView = textureImageView

	return nil
}

func (v *VulkanRenderer) beginSingleTimeCommands() (vk.CommandBuffer, error) {
	cbai := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		Level:              vk.CommandBufferLevelPrimary,
		CommandPool:        v.commandPool,
		CommandBufferCount: 1,
	}

	var commandBuffer vk.CommandBuffer
	if err := vk.Error(vk.AllocateCommandBuffers(v.logicalDevice, &cbai, []vk.CommandBuffer{commandBuffer})); err != nil {
		return nil, fmt.Errorf("vk.AllocateCommandBuffers(): %s", err.Error())
	}

	cbbi := vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	}

	if err := vk.Error(vk.BeginCommandBuffer(commandBuffer, &cbbi)); err != nil {
		vk.FreeCommandBuffers(v.logicalDevice, v.commandPool, 1, []vk.CommandBuffer{commandBuffer})
		return nil, fmt.Errorf("vk.BeginCommandBuffer(): %s", err.Error())
	}

	return commandBuffer, nil
}

func (v *VulkanRenderer) endSingleTimeCommands(commandBuffer vk.CommandBuffer) error {
	if err := vk.Error(vk.EndCommandBuffer(commandBuffer)); err != nil {
		return fmt.Errorf("vk.EndCommandBuffer(): %s", err.Error())
	}

	si := vk.SubmitInfo{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{commandBuffer},
	}

	if err := vk.Error(vk.QueueSubmit(v.deviceQueue, 1, []vk.SubmitInfo{si}, nil)); err != nil {
		return fmt.Errorf("vk.QueueSubmit(): %s", err.Error())
	}

	vk.QueueWaitIdle(v.deviceQueue)

	vk.FreeCommandBuffers(v.logicalDevice, v.commandPool, 1, []vk.CommandBuffer{commandBuffer})
	return nil
}

func (v *VulkanRenderer) transitionLayout(img vk.Image, format vk.Format, old vk.ImageLayout, new vk.ImageLayout) error {
	cmd, err := v.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	barrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           old,
		NewLayout:           new,
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               img,
		SubresourceRange: vk.ImageSubresourceRange{
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}

	var srcStage, dstStage vk.PipelineStageFlags
	if old == vk.ImageLayoutUndefined && new == vk.ImageLayoutTransferDstOptimal {
		barrier.SrcAccessMask = 0
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
	} else if old == vk.ImageLayoutTransferDstOptimal && new == vk.ImageLayoutShaderReadOnlyOptimal {
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)
		srcStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	} else {
		return fmt.Errorf("unsupported layout transition")
	}

	vk.CmdPipelineBarrier(cmd, srcStage, dstStage, 0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{barrier})

	if err := v.endSingleTimeCommands(cmd); err != nil {
		return err
	}
	return nil
}

func (v *VulkanRenderer) copyBuffer(src, dst vk.Buffer, size vk.DeviceSize) error {
	cmd, err := v.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	bc := vk.BufferCopy{
		Size: size,
	}
	vk.CmdCopyBuffer(cmd, src, dst, 1, []vk.BufferCopy{bc})

	if err := v.endSingleTimeCommands(cmd); err != nil {
		return err
	}
	return nil
}

func (v *VulkanRenderer) copyBufferToImage(buf vk.Buffer, img vk.Image, width, height uint32) error {
	cmd, err := v.beginSingleTimeCommands()
	if err != nil {
		return err
	}

	bic := vk.BufferImageCopy{
		ImageOffset: vk.Offset3D{},
		ImageExtent: vk.Extent3D{
			Height: height,
			Width:  width,
		},
	}
	vk.CmdCopyBufferToImage(cmd, buf, img, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{bic})

	if err := v.endSingleTimeCommands(cmd); err != nil {
		return err
	}
	return nil
}

func (v *VulkanRenderer) createUniformBuffers() error {
	bufferSize := vk.DeviceSize(unsafe.Sizeof(model.Uniform{}))
	uniformBuffers := make([]vk.Buffer, len(v.swapchainImages))
	uniformBuffersMemory := make([]vk.DeviceMemory, len(v.swapchainImages))

	for idx := 0; idx < len(v.swapchainImages); idx++ {
		bci := vk.BufferCreateInfo{
			SType:       vk.StructureTypeBufferCreateInfo,
			Size:        bufferSize,
			Usage:       vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit),
			SharingMode: vk.SharingModeExclusive,
		}
		if err := vk.Error(vk.CreateBuffer(v.logicalDevice, &bci, nil, &uniformBuffers[idx])); err != nil {
			return fmt.Errorf("vk.CreateBuffer(): %s", err.Error())
		}

		var memoryRequirements vk.MemoryRequirements
		vk.GetBufferMemoryRequirements(v.logicalDevice, uniformBuffers[idx], &memoryRequirements)
		memoryRequirements.Deref()
		memTypeIdx, err := findMemoryType(v.physicalDevice, memoryRequirements.MemoryTypeBits, vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
		if err != nil {
			return err
		}

		mai := vk.MemoryAllocateInfo{
			SType:           vk.StructureTypeMemoryAllocateInfo,
			AllocationSize:  memoryRequirements.Size,
			MemoryTypeIndex: memTypeIdx,
		}

		if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &mai, nil, &uniformBuffersMemory[idx])); err != nil {
			return fmt.Errorf("vk.AllocateMemory(): %s", err.Error())
		}

		vk.BindBufferMemory(v.logicalDevice, uniformBuffers[idx], uniformBuffersMemory[idx], 0)
	}

	v.uniformBuffers = uniformBuffers
	v.uniformBuffersMemory = uniformBuffersMemory
	return nil
}

func (v *VulkanRenderer) createVertexBuffers() error {
	bci := vk.BufferCreateInfo{
		SType:       vk.StructureTypeBufferCreateInfo,
		Size:        vk.DeviceSize(int(unsafe.Sizeof(model.Vertex{})) * len(v.vertices)),
		Usage:       vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit),
		SharingMode: vk.SharingModeExclusive,
	}

	var vertexBuffer vk.Buffer
	if err := vk.Error(vk.CreateBuffer(v.logicalDevice, &bci, nil, &vertexBuffer)); err != nil {
		return fmt.Errorf("vk.CreateBuffer(): %s", err.Error())
	}
	v.vertexBuffer = vertexBuffer

	memoryRequirements := vk.MemoryRequirements{}
	vk.GetBufferMemoryRequirements(v.logicalDevice, vertexBuffer, &memoryRequirements)
	memoryRequirements.Deref()

	memoryTypeIndex, err := findMemoryType(v.physicalDevice, memoryRequirements.MemoryTypeBits, vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	mai := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memoryRequirements.Size,
		MemoryTypeIndex: memoryTypeIndex,
	}

	var vertexMemory vk.DeviceMemory
	if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &mai, nil, &vertexMemory)); err != nil {
		return fmt.Errorf("vk.AllocateMemory(): %s", err.Error())
	}
	v.vertexMemory = vertexMemory

	if err := vk.Error(vk.BindBufferMemory(v.logicalDevice, vertexBuffer, vertexMemory, 0)); err != nil {
		return fmt.Errorf("vk.BindBufferMemory(): %s", err.Error())
	}

	var vertexMappedMemory unsafe.Pointer
	vk.MapMemory(v.logicalDevice, vertexMemory, 0, bci.Size, 0, &vertexMappedMemory)

	vertexCastMemory := *(*[]model.Vertex)(unsafe.Pointer(&sliceHeader{
		Data: uintptr(vertexMappedMemory),
		Cap:  len(v.vertices),
		Len:  len(v.vertices),
	}))
	copy(vertexCastMemory, v.vertices[:])

	// Approach #2
	// buf := new(bytes.Buffer)
	// binary.Write(buf, binary.LittleEndian, v.vertices)
	// rawBytes := buf.Bytes()
	// vk.Memcopy(vertexMappedMemory, rawBytes)

	vk.UnmapMemory(v.logicalDevice, vertexMemory)

	return nil
}

func findMemoryType(device vk.PhysicalDevice, filter uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	memoryProperties := vk.PhysicalDeviceMemoryProperties{}
	vk.GetPhysicalDeviceMemoryProperties(device, &memoryProperties)
	memoryProperties.Deref()

	for idx := uint32(0); idx < memoryProperties.MemoryTypeCount; idx++ {
		memoryProperties.MemoryTypes[idx].Deref()
		if filter&(1<<idx) != 0 && (memoryProperties.MemoryTypes[idx].PropertyFlags&properties) == properties {
			return idx, nil
		}
	}
	return 0, errors.New("memory type not found for vertex buffer")
}

func (v *VulkanRenderer) destroyBeforeRecreatePipeline() {
	vk.FreeCommandBuffers(v.logicalDevice, v.commandPool, uint32(len(v.commandBuffers)), v.commandBuffers)

	for _, mem := range v.uniformBuffersMemory {
		vk.FreeMemory(v.logicalDevice, mem, nil)
	}

	for _, buf := range v.uniformBuffers {
		vk.DestroyBuffer(v.logicalDevice, buf, nil)
	}

	for _, fb := range v.framebuffers {
		vk.DestroyFramebuffer(v.logicalDevice, fb, nil)
	}
	v.framebuffers = []vk.Framebuffer{}

	vk.DestroyDescriptorPool(v.logicalDevice, v.descriptorPool, nil)
	for _, descriptorLayout := range v.descriptorSetLayouts {
		vk.DestroyDescriptorSetLayout(v.logicalDevice, descriptorLayout, nil)
	}

	// Swapchain resources
	for _, iv := range v.swapchainImageViews {
		vk.DestroyImageView(v.logicalDevice, iv, nil)
	}
	v.swapchainImageViews = []vk.ImageView{}

	for _, i := range v.swapchainImages {
		vk.DestroyImage(v.logicalDevice, i, nil)
	}
	v.swapchainImages = []vk.Image{}

	// Depth image resources
	vk.DestroyImage(v.logicalDevice, v.depthImage, nil)
	vk.DestroyImageView(v.logicalDevice, v.depthImageView, nil)
	vk.FreeMemory(v.logicalDevice, v.depthImageMemory, nil)

	vk.DestroyRenderPass(v.logicalDevice, v.renderPass, nil)

	vk.DestroyPipeline(v.logicalDevice, v.pipeline, nil)
	vk.DestroyPipelineLayout(v.logicalDevice, v.pipelineLayout, nil)
}

func (v *VulkanRenderer) recreatePipeline() error {
	vk.DeviceWaitIdle(v.logicalDevice)
	v.destroyBeforeRecreatePipeline()

	if err := v.createSwapchain(v.swapchain); err != nil {
		return err
	}

	if err := v.prepareDepthImage(); err != nil {
		return err
	}

	if err := v.createRenderPass(); err != nil {
		return err
	}

	if err := v.createPipelineLayout(); err != nil {
		return err
	}

	if err := v.createPipeline(); err != nil {
		return err
	}

	if err := v.createImageViews(); err != nil {
		return err
	}

	if err := v.createFramebuffers(); err != nil {
		return err
	}

	if err := v.prepareDescriptorPool(); err != nil {
		return err
	}

	if err := v.createUniformBuffers(); err != nil {
		return err
	}

	if err := v.createDescriptorSets(); err != nil {
		return err
	}

	if err := v.allocateCommandBuffers(); err != nil {
		return err
	}

	if err := v.buildCommandBuffers(); err != nil {
		return err
	}

	// prepareDescriptorPool (later)
	return nil
}

func (v *VulkanRenderer) createPipelineCache() error {
	pcci := vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}

	var pipelineCache vk.PipelineCache
	if err := vk.Error(vk.CreatePipelineCache(v.logicalDevice, &pcci, nil, &pipelineCache)); err != nil {
		return errors.New("vk.CreatePipelineCache(): " + err.Error())
	}
	v.pipelineCache = pipelineCache
	return nil
}

func (v *VulkanRenderer) buildCommandBuffers() error {
	for idx, commandBuffer := range v.commandBuffers {
		cbbi := vk.CommandBufferBeginInfo{
			SType: vk.StructureTypeCommandBufferBeginInfo,
			Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
		}
		if err := vk.Error(vk.BeginCommandBuffer(commandBuffer, &cbbi)); err != nil {
			return fmt.Errorf("vk.BeginCommandBuffer()[%d]: %s", idx, err.Error())
		}

		clearValues := make([]vk.ClearValue, 2)
		clearValues[1].SetDepthStencil(1, 0)
		clearValues[0].SetColor([]float32{
			0.005, 0.005, 0.005, 0.005,
		})

		rpbi := vk.RenderPassBeginInfo{
			SType:       vk.StructureTypeRenderPassBeginInfo,
			RenderPass:  v.renderPass,
			Framebuffer: v.framebuffers[idx],
			RenderArea: vk.Rect2D{
				Offset: vk.Offset2D{
					X: 0, Y: 0,
				},
				Extent: vk.Extent2D{
					Width:  v.currentSurfaceWidth,
					Height: v.currentSurfaceHeight,
				},
			},
			ClearValueCount: 2,
			PClearValues:    clearValues,
		}
		vk.CmdBeginRenderPass(commandBuffer, &rpbi, vk.SubpassContentsInline)
		vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, v.pipeline)
		vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{v.viewport})
		vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{v.scissor})
		vk.CmdBindVertexBuffers(commandBuffer, 0, 1, []vk.Buffer{v.vertexBuffer}, []vk.DeviceSize{0})
		vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, v.pipelineLayout, 0, 1, v.descriptorSets, 0, nil)
		vk.CmdDraw(commandBuffer, uint32(len(v.vertices)), 1, 0, 0)
		vk.CmdEndRenderPass(commandBuffer)

		if err := vk.Error(vk.EndCommandBuffer(commandBuffer)); err != nil {
			return fmt.Errorf("vk.EndCommandBuffer()[%d]: %s", idx, err.Error())
		}
	}
	return nil
}

var constant float32

func (v *VulkanRenderer) updateUniformBuffers(imageIdx uint32) {
	constant += 0.005
	ubo := model.Uniform{
		Model:      glm.HomogRotate3D(constant, glm.Vec3{0, 1, 0}),
		View:       glm.LookAt(2, 2, 2, 0, 0, 0, 0, 0, 1),
		Projection: glm.Perspective(45, (float32)(v.currentSurfaceWidth)/(float32)(v.currentSurfaceHeight), 0.1, 10),
	}
	ubo.Projection[5] *= -1 // Flip from OpenGl to Vulkan projection

	var mappedMemory unsafe.Pointer
	vk.MapMemory(v.logicalDevice, v.uniformBuffersMemory[imageIdx], 0, vk.DeviceSize(unsafe.Sizeof(ubo)), 0, &mappedMemory)
	castMemory := *(*[]model.Uniform)(unsafe.Pointer(&sliceHeader{
		Data: uintptr(mappedMemory),
		Cap:  1,
		Len:  1,
	}))
	copy(castMemory, []model.Uniform{ubo})
	vk.UnmapMemory(v.logicalDevice, v.uniformBuffersMemory[imageIdx])
}

// Draw implements interface
func (v *VulkanRenderer) Draw() error {
	vk.WaitForFences(v.logicalDevice, 1, []vk.Fence{v.imageFence}, 0, math.MaxUint32)
	vk.ResetFences(v.logicalDevice, 1, []vk.Fence{v.imageFence})

	if result := vk.AcquireNextImage(v.logicalDevice, v.swapchain, math.MaxUint64, v.imageAvailableSemaphore, nil, &v.imageIndex); result == vk.ErrorOutOfDate {
		if err := v.recreatePipeline(); err != nil {
			return err
		}
		return nil
	}

	v.updateUniformBuffers(v.imageIndex)

	submit := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    []vk.Semaphore{v.imageAvailableSemaphore},
		PWaitDstStageMask: []vk.PipelineStageFlags{
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		},
		CommandBufferCount:   1,
		PCommandBuffers:      []vk.CommandBuffer{v.commandBuffers[v.imageIndex]},
		SignalSemaphoreCount: 1,
		PSignalSemaphores:    []vk.Semaphore{v.renderFinishedSemphore},
	}}

	if err := vk.Error(vk.QueueSubmit(v.deviceQueue, 1, submit, v.imageFence)); err != nil {
		return err
	}

	presentInfo := vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    []vk.Semaphore{v.renderFinishedSemphore},
		SwapchainCount:     1,
		PSwapchains:        []vk.Swapchain{v.swapchain},
		PImageIndices:      []uint32{v.imageIndex},
	}

	presentResult := vk.QueuePresent(v.deviceQueue, &presentInfo)
	if presentResult == vk.ErrorOutOfDate {
		if err := v.recreatePipeline(); err != nil {
			return err
		}
		return nil
	}

	if err := vk.Error(presentResult); err != nil {
		return errors.New("vk.QueuePresent(): " + err.Error())
	}

	return nil
}

// Present implements interface
func (v *VulkanRenderer) Present() error {
	return nil
}

func (v *VulkanRenderer) createViewport() {
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(v.currentSurfaceWidth),
		Height:   float32(v.currentSurfaceHeight),
		MinDepth: 0,
		MaxDepth: 1,
	}

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  v.currentSurfaceWidth,
			Height: v.currentSurfaceHeight,
		},
	}
	v.viewport = viewport
	v.scissor = scissor
}

func (v *VulkanRenderer) createDescriptorSets() error {
	descriptorSets := make([]vk.DescriptorSet, len(v.swapchainImages))
	dsai := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     v.descriptorPool,
		DescriptorSetCount: 1,
		PSetLayouts:        v.descriptorSetLayouts,
	}

	for idx := range v.swapchainImages {
		if err := vk.Error(vk.AllocateDescriptorSets(v.logicalDevice, &dsai, &descriptorSets[idx])); err != nil {
			return fmt.Errorf("vk.AllocateDescriptorSets(): %s", err.Error())
		}

		dbi := vk.DescriptorBufferInfo{
			Buffer: v.uniformBuffers[idx],
			Offset: 0,
			Range:  vk.DeviceSize(unsafe.Sizeof(model.Uniform{})),
		}
		dii := vk.DescriptorImageInfo{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   v.textureImageView,
			Sampler:     v.textureSampler,
		}
		wds := []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          descriptorSets[idx],
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			DescriptorCount: 1,
			PBufferInfo:     []vk.DescriptorBufferInfo{dbi},
		}, vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          descriptorSets[idx],
			DstBinding:      1,
			DstArrayElement: 0,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: 1,
			PBufferInfo:     []vk.DescriptorBufferInfo{dbi},
			PImageInfo:      []vk.DescriptorImageInfo{dii},
		}}
		vk.UpdateDescriptorSets(v.logicalDevice, uint32(len(wds)), wds, 0, nil)
	}
	v.descriptorSets = descriptorSets
	return nil
}

func (v *VulkanRenderer) prepareDescriptorPool() error {
	dpci := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       uint32(len(v.swapchainImages)),
		PoolSizeCount: 1,
		PPoolSizes: []vk.DescriptorPoolSize{
			{
				Type:            vk.DescriptorTypeUniformBuffer,
				DescriptorCount: uint32(len(v.swapchainImages)),
			},
			{
				Type:            vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount: uint32(len(v.swapchainImages)),
			}},
	}

	var descriptorPool vk.DescriptorPool
	if err := vk.Error(vk.CreateDescriptorPool(v.logicalDevice, &dpci, nil, &descriptorPool)); err != nil {
		return err
	}
	v.descriptorPool = descriptorPool

	return nil
}

func (v *VulkanRenderer) prepareDepthImage() error {
	depthFormat := vk.FormatD16Unorm
	ici := vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    depthFormat,
		Extent: vk.Extent3D{
			Width:  v.currentSurfaceWidth,
			Height: v.currentSurfaceHeight,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
	}

	var image vk.Image
	if err := vk.Error(vk.CreateImage(v.logicalDevice, &ici, nil, &image)); err != nil {
		return err
	}

	var memoryRequirements vk.MemoryRequirements
	vk.GetImageMemoryRequirements(v.logicalDevice, image, &memoryRequirements)
	memoryRequirements.Deref()

	memoryType, err := v.getMemoryType(memoryRequirements.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit))
	if err != nil {
		return err
	}

	mai := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memoryRequirements.Size,
		MemoryTypeIndex: memoryType,
	}

	var memory vk.DeviceMemory
	if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &mai, nil, &memory)); err != nil {
		return err
	}

	if err := vk.Error(vk.BindImageMemory(v.logicalDevice, image, memory, 0)); err != nil {
		return err
	}

	ivci := vk.ImageViewCreateInfo{
		SType:  vk.StructureTypeImageViewCreateInfo,
		Format: depthFormat,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			LevelCount: 1,
			LayerCount: 1,
		},
		ViewType: vk.ImageViewType2d,
		Image:    image,
	}

	var view vk.ImageView
	if err := vk.Error(vk.CreateImageView(v.logicalDevice, &ivci, nil, &view)); err != nil {
		return err
	}

	v.depthImage = image
	v.depthImageView = view
	v.depthImageMemory = memory
	v.depthImageFormat = depthFormat

	return nil
}

func (v *VulkanRenderer) createSynchronization() error {
	sci := vk.SemaphoreCreateInfo{
		SType: vk.StructureTypeSemaphoreCreateInfo,
	}
	fci := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
		Flags: vk.FenceCreateFlags(vk.FenceCreateSignaledBit),
	}

	var (
		imageAvailableSemaphore vk.Semaphore
		renderFinishedSemphore  vk.Semaphore
		fence                   vk.Fence
	)

	if err := vk.Error(vk.CreateSemaphore(v.logicalDevice, &sci, nil, &imageAvailableSemaphore)); err != nil {
		return errors.New("vk.CreateSemaphore(): " + err.Error())
	}
	if err := vk.Error(vk.CreateSemaphore(v.logicalDevice, &sci, nil, &renderFinishedSemphore)); err != nil {
		return errors.New("vk.CreateSemaphore(): " + err.Error())
	}
	if err := vk.Error(vk.CreateFence(v.logicalDevice, &fci, nil, &fence)); err != nil {
		return errors.New("vk.CreateFence(): " + err.Error())
	}

	v.imageAvailableSemaphore = imageAvailableSemaphore
	v.renderFinishedSemphore = renderFinishedSemphore
	v.imageFence = fence

	return nil
}

func (v *VulkanRenderer) createCommandPool() error {
	cpci := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: v.graphicsQueueIndex,
	}

	var commandPool vk.CommandPool
	if err := vk.Error(vk.CreateCommandPool(v.logicalDevice, &cpci, nil, &commandPool)); err != nil {
		return errors.New("vk.CreateCommandPool(): " + err.Error())
	}
	v.commandPool = commandPool

	return nil
}

func (v *VulkanRenderer) allocateCommandBuffers() error {

	cbai := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        v.commandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: uint32(len(v.swapchainImageViews)),
	}

	commandBuffers := make([]vk.CommandBuffer, len(v.swapchainImageViews))
	if err := vk.Error(vk.AllocateCommandBuffers(v.logicalDevice, &cbai, commandBuffers)); err != nil {
		return errors.New("vk.AllocateCommandBuffers(): " + err.Error())
	}
	v.commandBuffers = commandBuffers

	return nil
}

func (v *VulkanRenderer) prepareUniformBuffers() error {
	// vk.CreateBuffer
	// vk.AllocateMemory
	// vk.BindBufferMemory
	bci := vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		Size:  10,
		Usage: vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit),
	}

	var buffer vk.Buffer
	if err := vk.Error(vk.CreateBuffer(v.logicalDevice, &bci, nil, &buffer)); err != nil {
		return err
	}

	var memoryRequirements vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(v.logicalDevice, buffer, &memoryRequirements)
	memoryRequirements.Deref()

	memoryType, err := v.getMemoryType(memoryRequirements.MemoryTypeBits,
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return err
	}

	allocationInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memoryRequirements.Size,
		MemoryTypeIndex: memoryType,
	}

	var memory vk.DeviceMemory
	if err := vk.Error(vk.AllocateMemory(v.logicalDevice, &allocationInfo, nil, &memory)); err != nil {
		return err
	}

	if err := vk.Error(vk.BindBufferMemory(v.logicalDevice, buffer, memory, 0)); err != nil {
		return err
	}

	return nil
}

func (v *VulkanRenderer) createFramebuffers() error {
	for _, image := range v.swapchainImageViews {
		attachments := []vk.ImageView{
			image,
			v.depthImageView,
		}
		fci := vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      v.renderPass,
			AttachmentCount: uint32(len(attachments)),
			PAttachments:    attachments,
			Width:           v.currentSurfaceWidth,
			Height:          v.currentSurfaceHeight,
			Layers:          1,
		}

		var framebuffer vk.Framebuffer
		if err := vk.Error(vk.CreateFramebuffer(v.logicalDevice, &fci, nil, &framebuffer)); err != nil {
			return err
		}
		v.framebuffers = append(v.framebuffers, framebuffer)
	}
	return nil
}

func (v *VulkanRenderer) getMemoryType(typeBits uint32, properties vk.MemoryPropertyFlags) (uint32, error) {
	var memoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(v.physicalDevice, &memoryProperties)
	memoryProperties.Deref()

	for idx := uint32(0); idx < memoryProperties.MemoryTypeCount; idx++ {
		if (typeBits & 1) == 1 {
			memoryProperties.MemoryTypes[idx].Deref()
			if (memoryProperties.MemoryTypes[idx].PropertyFlags & properties) == properties {
				return idx, nil
			}
		}
		typeBits >>= 1
	}
	return 0, errors.New("requested memory type not found")
}

func (v *VulkanRenderer) createPipeline() error {
	pipelineShaderStagesInfo := make([]vk.PipelineShaderStageCreateInfo, len(v.shaders))
	for idx, shader := range v.shaders {

		var stage vk.ShaderStageFlagBits
		switch shader.Type() {
		case VertexShaderType:
			stage = vk.ShaderStageVertexBit
		case FragmentShaderType:
			stage = vk.ShaderStageFragmentBit
		default:
			return errors.New("unsupported shader type attempted creation")
		}

		var shaderModule vk.ShaderModule
		if sm, ok := shader.ShaderModule().(vk.ShaderModule); ok {
			shaderModule = sm
		} else {
			return errors.New("failed to assert shader module to it's original type")
		}

		pipelineShaderStagesInfo[idx].SType = vk.StructureTypePipelineShaderStageCreateInfo
		pipelineShaderStagesInfo[idx].Stage = stage
		pipelineShaderStagesInfo[idx].Module = shaderModule
		pipelineShaderStagesInfo[idx].PName = safeString("main")
	}

	vertexAttributeDescriptions := model.VertexAttributeDescriptions()
	vertexBindingDescriptions := model.VertexBindingDescriptions()

	gpci := []vk.GraphicsPipelineCreateInfo{{
		SType:      vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount: uint32(len(pipelineShaderStagesInfo)),
		PStages:    pipelineShaderStagesInfo,
		PVertexInputState: &vk.PipelineVertexInputStateCreateInfo{
			SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
			VertexAttributeDescriptionCount: uint32(len(vertexAttributeDescriptions)),
			PVertexAttributeDescriptions:    vertexAttributeDescriptions,
			VertexBindingDescriptionCount:   uint32(len(vertexBindingDescriptions)),
			PVertexBindingDescriptions:      vertexBindingDescriptions,
		},
		PInputAssemblyState: &vk.PipelineInputAssemblyStateCreateInfo{
			SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
			Topology: vk.PrimitiveTopologyTriangleList,
		},
		PViewportState: &vk.PipelineViewportStateCreateInfo{
			SType:         vk.StructureTypePipelineViewportStateCreateInfo,
			ViewportCount: 1,
			ScissorCount:  1,
		},
		PRasterizationState: &vk.PipelineRasterizationStateCreateInfo{
			SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
			PolygonMode: vk.PolygonModeFill,
			CullMode:    vk.CullModeFlags(vk.CullModeBackBit),
			FrontFace:   vk.FrontFaceClockwise,
			LineWidth:   1.0,
		},
		PDepthStencilState: &vk.PipelineDepthStencilStateCreateInfo{
			SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
			DepthTestEnable:       vk.True,
			DepthWriteEnable:      vk.True,
			DepthCompareOp:        vk.CompareOpLessOrEqual,
			DepthBoundsTestEnable: vk.False,
			Back: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
			StencilTestEnable: vk.False,
			Front: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
		},
		PMultisampleState: &vk.PipelineMultisampleStateCreateInfo{
			SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
			RasterizationSamples: vk.SampleCount1Bit,
		},
		PColorBlendState: &vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			AttachmentCount: 1,
			PAttachments: []vk.PipelineColorBlendAttachmentState{{
				ColorWriteMask: 0xF,
				BlendEnable:    vk.False,
			}},
		},
		PDynamicState: &vk.PipelineDynamicStateCreateInfo{
			SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
			DynamicStateCount: 2,
			PDynamicStates: []vk.DynamicState{
				vk.DynamicStateScissor,
				vk.DynamicStateViewport,
			},
		},
		Layout:     v.pipelineLayout,
		RenderPass: v.renderPass,
	}}

	pipelines := make([]vk.Pipeline, len(gpci))
	if err := vk.Error(vk.CreateGraphicsPipelines(v.logicalDevice, v.pipelineCache, uint32(len(gpci)), gpci, nil, pipelines)); err != nil {
		return errors.New("vk.CreateGraphicsPipelines(): " + err.Error())
	}
	v.pipeline = pipelines[0]
	return nil
}

func (v *VulkanRenderer) createSwapchain(oldSwapchain vk.Swapchain) error {
	var surfaceCapabilities vk.SurfaceCapabilities
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(v.physicalDevice, v.surface, &surfaceCapabilities)); err != nil {
		return errors.New("vk.GetPhysicalDeviceSurfaceCapabilities(): " + err.Error())
	}

	// In case swapchain is being recreated
	if oldSwapchain != nil {
		surfaceCapabilities.Deref()
		surfaceCapabilities.CurrentExtent.Deref()
		v.currentSurfaceHeight = surfaceCapabilities.CurrentExtent.Height
		v.currentSurfaceWidth = surfaceCapabilities.CurrentExtent.Width
	}

	compositeAlpha := vk.CompositeAlphaOpaqueBit
	compositeAlphaFlags := []vk.CompositeAlphaFlagBits{
		vk.CompositeAlphaOpaqueBit,
		vk.CompositeAlphaPreMultipliedBit,
		vk.CompositeAlphaPostMultipliedBit,
		vk.CompositeAlphaInheritBit,
	}

	// CompositeAlpha
	for i := 0; i < len(compositeAlphaFlags); i++ {
		alphaFlags := vk.CompositeAlphaFlags(compositeAlphaFlags[i])
		flagSupported := surfaceCapabilities.SupportedCompositeAlpha&alphaFlags != 0
		if flagSupported {
			compositeAlpha = compositeAlphaFlags[i]
			break
		}
	}

	var swapchain vk.Swapchain
	scci := vk.SwapchainCreateInfo{
		SType:           vk.StructureTypeSwapchainCreateInfo,
		Surface:         v.surface,
		MinImageCount:   v.configuration.SwapchainSize,
		ImageFormat:     v.imageFormat,
		ImageColorSpace: v.imageColorspace,
		ImageExtent: vk.Extent2D{
			Width:  v.currentSurfaceWidth,
			Height: v.currentSurfaceHeight,
		},
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     vk.SurfaceTransformIdentityBit,
		CompositeAlpha:   compositeAlpha,
		PresentMode:      vk.PresentModeFifo,
		Clipped:          vk.True,
		ImageArrayLayers: 1,
		ImageSharingMode: vk.SharingModeExclusive,
		OldSwapchain:     oldSwapchain,
	}

	if err := vk.Error(vk.CreateSwapchain(v.logicalDevice, &scci, nil, &swapchain)); err != nil {
		return errors.New("vk.CreateSwapchain(): " + err.Error())
	}
	v.swapchain = swapchain

	var numImages uint32
	if err := vk.Error(vk.GetSwapchainImages(v.logicalDevice, v.swapchain, &numImages, nil)); err != nil {
		return errors.New("vk.GetSwapchainImages(num): " + err.Error())
	}

	v.swapchainImages = make([]vk.Image, numImages)
	if err := vk.Error(vk.GetSwapchainImages(v.logicalDevice, v.swapchain, &numImages, v.swapchainImages)); err != nil {
		return errors.New("vk.GetSwapchainImages(images): " + err.Error())
	}
	return nil
}

func (v *VulkanRenderer) createPipelineLayout() error {
	dslci := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 1,
		PBindings: []vk.DescriptorSetLayoutBinding{
			vk.DescriptorSetLayoutBinding{
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageVertexBit),
				Binding:         0,
			},
			vk.DescriptorSetLayoutBinding{
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
				Binding:         1,
			},
		},
	}

	var descriptorSetLayouts []vk.DescriptorSetLayout
	for range v.swapchainImages {
		var descriptorSetLayout vk.DescriptorSetLayout
		if err := vk.Error(vk.CreateDescriptorSetLayout(v.logicalDevice, &dslci, nil, &descriptorSetLayout)); err != nil {
			return errors.New("vk.CreateDescriptorSetLayout(): " + err.Error())
		}
		descriptorSetLayouts = append(descriptorSetLayouts, descriptorSetLayout)
	}
	v.descriptorSetLayouts = descriptorSetLayouts

	plci := vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: uint32(len(v.descriptorSetLayouts)),
		PSetLayouts:    v.descriptorSetLayouts,
	}

	var pipelineLayout vk.PipelineLayout
	if err := vk.Error(vk.CreatePipelineLayout(v.logicalDevice, &plci, nil, &pipelineLayout)); err != nil {
		return errors.New("vk.CreatePipelineLayout(): " + err.Error())
	}
	v.pipelineLayout = pipelineLayout
	return nil
}

func (v *VulkanRenderer) createRenderPass() error {
	swapchainAttachments := []vk.AttachmentDescription{
		vk.AttachmentDescription{
			Format:         v.imageFormat,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutPresentSrc,
		},
		vk.AttachmentDescription{
			Format:         vk.FormatD16Unorm,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpDontCare,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		},
	}

	colorAttachmentRef := []vk.AttachmentReference{{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}}

	subpassDependency := vk.SubpassDependency{
		SrcSubpass:    vk.SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		SrcAccessMask: 0,
		DstStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit),
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: uint32(len(colorAttachmentRef)),
		PColorAttachments:    colorAttachmentRef,
	}

	rpci := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(swapchainAttachments)),
		PAttachments:    swapchainAttachments,
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{subpassDependency},
	}

	var renderPass vk.RenderPass
	if err := vk.Error(vk.CreateRenderPass(v.logicalDevice, &rpci, nil, &renderPass)); err != nil {
		return errors.New("vk.CreateRenderPass(): " + err.Error())
	}
	v.renderPass = renderPass
	v.swapchainAttachments = swapchainAttachments
	return nil
}

func (v *VulkanRenderer) createImageViews() error {
	for idx := 0; idx < len(v.swapchainImages); idx++ {
		var imageView vk.ImageView
		ivci := vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    v.swapchainImages[idx],
			ViewType: vk.ImageViewType2d,
			Format:   v.imageFormat,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleIdentity,
				G: vk.ComponentSwizzleIdentity,
				B: vk.ComponentSwizzleIdentity,
				A: vk.ComponentSwizzleIdentity,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     1,
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}

		if err := vk.Error(vk.CreateImageView(v.logicalDevice, &ivci, nil, &imageView)); err != nil {
			return errors.New("with index " + string(idx) + "vk.CreateImageView(): " + err.Error())
		}

		v.swapchainImageViews = append(v.swapchainImageViews, imageView)
	}
	return nil
}

func (v *VulkanRenderer) loadShaders() error {
	var shaders []Shader
	shaderFiles, shaderTypes, err := loadShaderFilesFromDirectory(v.configuration.ShaderDirectory)
	if err != nil {
		return err
	}

	for idx, val := range shaderFiles {
		shader, err := NewVulkanShader(val, shaderTypes[idx], v.logicalDevice)
		if err != nil {
			return err
		}
		shaders = append(shaders, shader)
	}
	v.shaders = shaders
	return nil
}

// DeviceIsSuitable implements interface
func (v VulkanRenderer) DeviceIsSuitable(device vk.PhysicalDevice) (bool, string) {
	// TODO: Add device suitability checking
	return true, ""
}

// Destroy implements interface
func (v *VulkanRenderer) Destroy() {
	vk.DeviceWaitIdle(v.logicalDevice)

	for _, shader := range v.shaders {
		shader.Destroy()
	}

	vk.DestroySemaphore(v.logicalDevice, v.imageAvailableSemaphore, nil)
	vk.DestroySemaphore(v.logicalDevice, v.renderFinishedSemphore, nil)
	vk.DestroyFence(v.logicalDevice, v.imageFence, nil)

	vk.DestroyCommandPool(v.logicalDevice, v.commandPool, nil)

	for _, mem := range v.uniformBuffersMemory {
		vk.FreeMemory(v.logicalDevice, mem, nil)
	}

	for _, buf := range v.uniformBuffers {
		vk.DestroyBuffer(v.logicalDevice, buf, nil)
	}

	for _, f := range v.framebuffers {
		vk.DestroyFramebuffer(v.logicalDevice, f, nil)
	}

	for _, i := range v.swapchainImageViews {
		vk.DestroyImageView(v.logicalDevice, i, nil)
	}

	vk.DestroyDescriptorPool(v.logicalDevice, v.descriptorPool, nil)
	for _, descriptorLayout := range v.descriptorSetLayouts {
		vk.DestroyDescriptorSetLayout(v.logicalDevice, descriptorLayout, nil)
	}

	vk.DestroyPipeline(v.logicalDevice, v.pipeline, nil)
	vk.DestroyPipelineCache(v.logicalDevice, v.pipelineCache, nil)
	vk.DestroyRenderPass(v.logicalDevice, v.renderPass, nil)
	vk.DestroyPipelineLayout(v.logicalDevice, v.pipelineLayout, nil)

	vk.DestroyBuffer(v.logicalDevice, v.vertexBuffer, nil)
	vk.FreeMemory(v.logicalDevice, v.vertexMemory, nil)

	vk.FreeMemory(v.logicalDevice, v.depthImageMemory, nil)
	vk.DestroyImageView(v.logicalDevice, v.depthImageView, nil)
	vk.DestroyImage(v.logicalDevice, v.depthImage, nil)

	vk.FreeMemory(v.logicalDevice, v.textureMemory, nil)
	vk.DestroyBuffer(v.logicalDevice, v.textureBuffer, nil)
	vk.FreeMemory(v.logicalDevice, v.textureImageMemory, nil)
	vk.DestroyImageView(v.logicalDevice, v.textureImageView, nil)
	vk.DestroyImage(v.logicalDevice, v.textureImage, nil)

	vk.DestroySampler(v.logicalDevice, v.textureSampler, nil)

	vk.DestroySwapchain(v.logicalDevice, v.swapchain, nil)
	vk.DestroyDevice(v.logicalDevice, nil)
}

// NewVulkanShader creates a Vulkan specific shader wrapper
func NewVulkanShader(path string, shaderType ShaderType, device vk.Device) (Shader, error) {
	splitPath := strings.Split(path, "/")
	filename := splitPath[len(splitPath)-1]
	shaderName := strings.Split(filename, ".")[0]

	shaderContents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	smci := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(shaderContents)),
		PCode:    sliceUint32(shaderContents),
	}

	var shader vk.ShaderModule
	if err := vk.Error(vk.CreateShaderModule(device, &smci, nil, &shader)); err != nil {
		return nil, fmt.Errorf("vk.CreateShaderModule(type %d): %s", shaderType, err.Error())
	}

	return &VulkanShader{
		shader:           shader,
		shaderType:       shaderType,
		shaderContents:   shaderContents,
		shaderCreateInfo: smci,
		name:             shaderName,
		device:           device,
	}, nil
}

// VulkanShader is a Vulkan specific shader
type VulkanShader struct {
	Destroyable
	Shader

	name             string
	shaderType       ShaderType
	device           vk.Device
	shader           vk.ShaderModule
	shaderContents   []byte
	slicedContents   []uint32
	shaderCreateInfo vk.ShaderModuleCreateInfo
}

// Type implements interface
func (v VulkanShader) Type() ShaderType {
	return v.shaderType
}

// ShaderModule is an accssor to the internal vk.ShaderModule
func (v VulkanShader) ShaderModule() interface{} {
	return v.shader
}

// Name implements interface
func (v VulkanShader) Name() string {
	return v.name
}

// Destroy implements interface
func (v VulkanShader) Destroy() {
	vk.DestroyShaderModule(v.device, v.shader, nil)
}
