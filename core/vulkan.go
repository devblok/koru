package core

import (
	"errors"
	"fmt"
	"io/ioutil"
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

// PhysicalDeviceInfo describes available physical properties of a rendering device
type PhysicalDeviceInfo struct {
	ID            int
	VendorID      int
	DriverVersion int
	Name          string
	Invalid       bool
	Extensions    []string
	Layers        []string
	Memory        vk.DeviceSize
	Features      vk.PhysicalDeviceFeatures
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
	for i := range pdi {
		pdi[i].Invalid = false
	}

	for i := 0; i < len(v.availableDevices); i++ {
		// Get device features
		vk.GetPhysicalDeviceFeatures(v.availableDevices[i], &pdi[i].Features)

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
			pdi[i].Memory = pdi[i].Memory + memoryProperties.MemoryHeaps[iMem].Size
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
		configuration:  cfg,
		surface:        instance.Surface(),
		physicalDevice: instance.AvailableDevices()[0],
	}, nil
}

// VulkanRenderer is a Vulkan API renderer
type VulkanRenderer struct {
	Destroyable
	Renderer

	configuration RendererConfiguration

	surface             vk.Surface
	swapchain           vk.Swapchain
	swapchainImages     []vk.Image
	logicalDevice       vk.Device
	physicalDevice      vk.PhysicalDevice
	imageFormat         vk.Format
	imageColorspace     vk.ColorSpace
	imageViews          []vk.ImageView
	viewport            vk.Viewport
	scissor             vk.Rect2D
	pipelineLayout      vk.PipelineLayout
	descriptorSetLayout vk.DescriptorSetLayout
	renderPass          vk.RenderPass
	pipeline            vk.Pipeline
	currentQueueIndex   uint32
	graphicsQueueIndex  uint32
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
	if err := v.createSwapchain(); err != nil {
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

	/* Uniform buffers */
	if err := v.prepareUniformBuffers(); err != nil {
		return err
	}

	/* Pipeline Layout */
	if err := v.createPipelineLayout(); err != nil {
		return err
	}

	/* Render pass */
	if err := v.createRenderPass(); err != nil {
		return err
	}

	/* Shaders */
	shaders, err := v.loadShaders(v.configuration.ShaderDirectory)
	if err != nil {
		return err
	}

	/* Pipeline */
	if err := v.createPipeline(shaders); err != nil {
		return err
	}

	for _, shader := range shaders {
		shader.Destroy()
	}

	if err := v.createImageViews(); err != nil {
		return err
	}

	return nil
}

func (v *VulkanRenderer) createViewport() {
	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(v.configuration.ScreenWidth),
		Height:   float32(v.configuration.ScreenHeight),
		MinDepth: 0,
		MaxDepth: 1,
	}

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  v.configuration.ScreenWidth,
			Height: v.configuration.ScreenHeight,
		},
	}
	v.viewport = viewport
	v.scissor = scissor
}

func (v *VulkanRenderer) prepareDepthImage() error {
	depthFormat := vk.FormatD16Unorm
	ici := vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    depthFormat,
		Extent: vk.Extent3D{
			Width:  v.configuration.ScreenWidth,
			Height: v.configuration.ScreenHeight,
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

func (v *VulkanRenderer) createPipeline(shaders []Shader) error {
	pipelineShaderStagesInfo := make([]vk.PipelineShaderStageCreateInfo, len(shaders))
	for idx, shader := range shaders {

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

	// pcbas := []vk.PipelineColorBlendAttachmentState{{
	// 	ColorWriteMask:      0xF, // ColorComponentFlagBits -> R | G | B | A
	// 	BlendEnable:         vk.False,
	// 	SrcColorBlendFactor: vk.BlendFactorOne,
	// 	DstColorBlendFactor: vk.BlendFactorZero,
	// 	ColorBlendOp:        vk.BlendOpAdd,
	// 	SrcAlphaBlendFactor: vk.BlendFactorOne,
	// 	DstAlphaBlendFactor: vk.BlendFactorZero,
	// 	AlphaBlendOp:        vk.BlendOpZero,
	// }}

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

	pcci := vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}

	var pipelineCache vk.PipelineCache
	if err := vk.Error(vk.CreatePipelineCache(v.logicalDevice, &pcci, nil, &pipelineCache)); err != nil {
		return errors.New("vk.CreatePipelineCache(): " + err.Error())
	}

	pipelines := make([]vk.Pipeline, len(gpci))
	if err := vk.Error(vk.CreateGraphicsPipelines(v.logicalDevice, pipelineCache, uint32(len(gpci)), gpci, nil, pipelines)); err != nil {
		return errors.New("vk.CreateGraphicsPipelines(): " + err.Error())
	}
	v.pipeline = pipelines[0]
	return nil
}

func (v *VulkanRenderer) createSwapchain() error {
	var surfaceCapabilities vk.SurfaceCapabilities
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(v.physicalDevice, v.surface, &surfaceCapabilities)); err != nil {
		return errors.New("vk.GetPhysicalDeviceSurfaceCapabilities(): " + err.Error())
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
			Width:  v.configuration.ScreenWidth,
			Height: v.configuration.ScreenHeight,
		},
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:     vk.SurfaceTransformIdentityBit,
		CompositeAlpha:   compositeAlpha,
		PresentMode:      vk.PresentModeFifo,
		Clipped:          vk.True,
		ImageArrayLayers: 1,
		ImageSharingMode: vk.SharingModeExclusive,
		OldSwapchain:     nil,
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
	colorAttachment := vk.AttachmentDescription{
		Format:         v.imageFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}

	depthAttachment := vk.AttachmentDescription{
		Format:         vk.FormatD16Unorm,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpDontCare,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
	}

	colorAttachmentRef := []vk.AttachmentReference{{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: uint32(len(colorAttachmentRef)),
		PColorAttachments:    colorAttachmentRef,
	}

	rpci := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 2,
		PAttachments:    []vk.AttachmentDescription{colorAttachment, depthAttachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
	}

	var renderPass vk.RenderPass
	if err := vk.Error(vk.CreateRenderPass(v.logicalDevice, &rpci, nil, &renderPass)); err != nil {
		return errors.New("vk.CreateRenderPass(): " + err.Error())
	}
	v.renderPass = renderPass
	return nil
}

func (v VulkanRenderer) createImageViews() error {
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

		v.imageViews = append(v.imageViews, imageView)
	}
	return nil
}

func (v *VulkanRenderer) loadShaders(shaderDir string) ([]Shader, error) {
	var shaders []Shader
	shaderFiles, shaderTypes, err := loadShaderFilesFromDirectory(shaderDir)
	if err != nil {
		return nil, err
	}

	for idx, val := range shaderFiles {
		shader, err := NewVulkanShader(val, shaderTypes[idx], v.logicalDevice)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, shader)
	}
	return shaders, nil
}

// DeviceIsSuitable implements interface
func (v VulkanRenderer) DeviceIsSuitable(device vk.PhysicalDevice) (bool, string) {
	// TODO: Add device suitability checking
	return true, ""
}

// Destroy implements interface
func (v VulkanRenderer) Destroy() {
	vk.DestroyPipeline(v.logicalDevice, v.pipeline, nil)
	vk.DestroyRenderPass(v.logicalDevice, v.renderPass, nil)
	vk.DestroyPipelineLayout(v.logicalDevice, v.pipelineLayout, nil)

	for _, i := range v.imageViews {
		vk.DestroyImageView(v.logicalDevice, i, nil)
	}

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
