package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"unsafe"

	vk "github.com/vulkan-go/vulkan"
)

// DefaultVulkanApplicationInfo application info describes a Vulkan application
var DefaultVulkanApplicationInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 0, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "Koru3D\x00",
	PEngineName:        "Koru3D\x00",
}

// NewVulkanInstance creates a Vulkan instance
func NewVulkanInstance(appInfo *vk.ApplicationInfo, window unsafe.Pointer, cfg InstanceConfiguration) (Instance, error) {
	if cfg.DebugMode {
		cfg.Layers = append(cfg.Layers, "VK_LAYER_LUNARG_standard_validation\x00")
		cfg.Extensions = append(cfg.Extensions, "VK_EXT_debug_report\x00")
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
		PpEnabledExtensionNames: cfg.Extensions,
		EnabledLayerCount:       uint32(len(cfg.Layers)),
		PpEnabledLayerNames:     cfg.Layers,
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
	return &VulkanRenderer{
		configuration:        cfg,
		currentSurfaceHeight: cfg.ScreenHeight,
		currentSurfaceWidth:  cfg.ScreenWidth,
		surface:              instance.Surface(),
		physicalDevice:       instance.AvailableDevices()[0],
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

	descriptorSetLayout vk.DescriptorSetLayout
	descriptorPool      vk.DescriptorPool
	renderPass          vk.RenderPass

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
}

// Initialise implements interface
func (v *VulkanRenderer) Initialise() error {
	// TODO: Make extension name escaping bearable
	requiredExtensions := []string{
		vk.KhrSwapchainExtensionName + "\x00",
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
		PpEnabledExtensionNames: requiredExtensions,
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

	// if err := v.prepareDescriptorPool(); err != nil {
	// 	return err
	// }

	// if err := v.prepareDescriptorSet(); err != nil {
	// 	return err
	// }

	if err := v.createFramebuffers(); err != nil {
		return err
	}

	if err := v.createCommandPool(); err != nil {
		return err
	}

	if err := v.allocateCommandBuffers(); err != nil {
		return err
	}

	if err := v.createSynchronization(); err != nil {
		return err
	}

	/* Fill in command buffers */
	if err := v.buildCommandBuffers(); err != nil {
		return err
	}

	return nil
}

func (v *VulkanRenderer) destroyBeforeRecreatePipeline() {
	vk.FreeCommandBuffers(v.logicalDevice, v.commandPool, uint32(len(v.commandBuffers)), v.commandBuffers)
	v.commandBuffers = []vk.CommandBuffer{}
	vk.DestroyCommandPool(v.logicalDevice, v.commandPool, nil)

	for _, fb := range v.framebuffers {
		vk.DestroyFramebuffer(v.logicalDevice, fb, nil)
	}
	v.framebuffers = []vk.Framebuffer{}

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

	vk.DestroyPipeline(v.logicalDevice, v.pipeline, nil)
	vk.DestroyPipelineLayout(v.logicalDevice, v.pipelineLayout, nil)
	vk.DestroyDescriptorSetLayout(v.logicalDevice, v.descriptorSetLayout, nil)
	vk.DestroyRenderPass(v.logicalDevice, v.renderPass, nil)
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

	if err := v.createCommandPool(); err != nil {
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
			0.05, 0.05, 0.05, 0.05,
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
		vk.CmdDraw(commandBuffer, 3, 1, 0, 0)
		vk.CmdEndRenderPass(commandBuffer)

		if err := vk.Error(vk.EndCommandBuffer(commandBuffer)); err != nil {
			return fmt.Errorf("vk.EndCommandBuffer()[%d]: %s", idx, err.Error())
		}
	}
	return nil
}

// Draw implements interface
func (v *VulkanRenderer) Draw() error {
	if result := vk.AcquireNextImage(v.logicalDevice, v.swapchain, math.MaxUint64, v.imageAvailableSemaphore, nil, &v.imageIndex); result == vk.ErrorOutOfDate {
		if err := v.recreatePipeline(); err != nil {
			return err
		}
		return nil
	}

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

	if err := vk.Error(vk.QueueSubmit(v.deviceQueue, 1, submit, nil)); err != nil {
		return err
	}

	swapchains := []vk.Swapchain{v.swapchain}
	presentInfo := vk.PresentInfo{
		SType:              vk.StructureTypePresentInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    []vk.Semaphore{v.renderFinishedSemphore},
		SwapchainCount:     1,
		PSwapchains:        swapchains,
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

func (v *VulkanRenderer) prepareDescriptorPool() error {
	dpci := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       uint32(len(v.swapchainImages)),
		PoolSizeCount: 2,
		PPoolSizes: []vk.DescriptorPoolSize{{
			Type:            vk.DescriptorTypeUniformBuffer,
			DescriptorCount: uint32(len(v.swapchainImages)),
		}, {
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

func (v *VulkanRenderer) prepareDescriptorSet() error {
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
		pipelineShaderStagesInfo[idx].PName = "main\x00"
	}

	gpci := []vk.GraphicsPipelineCreateInfo{{
		SType:      vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount: uint32(len(pipelineShaderStagesInfo)),
		PStages:    pipelineShaderStagesInfo,
		PVertexInputState: &vk.PipelineVertexInputStateCreateInfo{
			SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
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
			},
		},
	}

	var descriptorSetLayout vk.DescriptorSetLayout
	if err := vk.Error(vk.CreateDescriptorSetLayout(v.logicalDevice, &dslci, nil, &descriptorSetLayout)); err != nil {
		return errors.New("vk.CreateDescriptorSetLayout(): " + err.Error())
	}
	v.descriptorSetLayout = descriptorSetLayout

	plci := vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts:    []vk.DescriptorSetLayout{v.descriptorSetLayout},
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

	for _, f := range v.framebuffers {
		vk.DestroyFramebuffer(v.logicalDevice, f, nil)
	}

	for _, i := range v.swapchainImageViews {
		vk.DestroyImageView(v.logicalDevice, i, nil)
	}

	vk.DestroyDescriptorPool(v.logicalDevice, v.descriptorPool, nil)
	vk.DestroyDescriptorSetLayout(v.logicalDevice, v.descriptorSetLayout, nil)

	vk.DestroyPipeline(v.logicalDevice, v.pipeline, nil)
	vk.DestroyPipelineCache(v.logicalDevice, v.pipelineCache, nil)
	vk.DestroyRenderPass(v.logicalDevice, v.renderPass, nil)
	vk.DestroyPipelineLayout(v.logicalDevice, v.pipelineLayout, nil)

	vk.FreeMemory(v.logicalDevice, v.depthImageMemory, nil)
	vk.DestroyImageView(v.logicalDevice, v.depthImageView, nil)
	vk.DestroyImage(v.logicalDevice, v.depthImage, nil)

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
