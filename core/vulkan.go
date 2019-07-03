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
	v := &VulkanInstance{
		configuration: cfg,
	}

	if v.configuration.DebugMode {
		v.configuration.Layers = append(v.configuration.Layers, "VK_LAYER_LUNARG_standard_validation\x00")
		v.configuration.Extensions = append(v.configuration.Extensions, "VK_EXT_debug_report\x00")
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

	{
		instanceInfo := vk.InstanceCreateInfo{
			SType:                   vk.StructureTypeInstanceCreateInfo,
			PApplicationInfo:        appInfo,
			EnabledExtensionCount:   uint32(len(v.configuration.Extensions)),
			PpEnabledExtensionNames: v.configuration.Extensions,
			EnabledLayerCount:       uint32(len(v.configuration.Layers)),
			PpEnabledLayerNames:     v.configuration.Layers,
		}

		var instance vk.Instance
		if err := vk.Error(vk.CreateInstance(&instanceInfo, nil, &instance)); err != nil {
			return nil, errors.New("vk.CreateInstance(): " + err.Error())
		}
		vk.InitInstance(instance)
		v.instance = instance
	}

	if physicalDevices, err := enumerateDevices(v.instance); err != nil {
		return nil, errors.New("core.enumerateDevices(): " + err.Error())
	} else {
		v.availableDevices = physicalDevices
	}

	return v, nil
}

// VulkanInstance describes a Vulkan API Instance
type VulkanInstance struct {
	Instance

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

// Extensions implements interface
func (v VulkanInstance) Extensions() []string {
	return v.configuration.Extensions
}

// AvailableDevices implements interface
func (v VulkanInstance) AvailableDevices() []vk.PhysicalDevice {
	return v.availableDevices
}

// Inner implements interface
func (v VulkanInstance) Inner() interface{} {
	return v.instance
}

// Destroy implements interface
func (v VulkanInstance) Destroy() {
	v.availableDevices = nil
	vk.DestroyInstance(v.instance, nil)
}

// NewVulkanRenderer creates a not yet initialised Vulkan API renderer
func NewVulkanRenderer(instance Instance, cfg RendererConfiguration) (Renderer, error) {
	return &VulkanRenderer{
		configuration: cfg,
		surface:       instance.Surface(),
		//physicalDeviceInfo: instance.PhysicalDevicesInfo()[0],
		physicalDevice: instance.AvailableDevices()[0],
	}, nil
}

// VulkanRenderer is a Vulkan API renderer
type VulkanRenderer struct {
	Renderer

	configuration RendererConfiguration

	surface            vk.Surface
	swapchain          vk.Swapchain
	swapchainImages    []vk.Image
	logicalDevice      vk.Device
	physicalDevice     vk.PhysicalDevice
	imageFormat        vk.Format
	imageViews         []vk.ImageView
	shaders            []vk.ShaderModule
	viewport           vk.Viewport
	scissor            vk.Rect2D
	pipelineLayout     vk.PipelineLayout
	renderPass         vk.RenderPass
	pipeline           vk.Pipeline
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
	v.logicalDevice = vkDevice

	/* Swapchain setup */
	var surfaceCapabilities vk.SurfaceCapabilities
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(v.physicalDevice, v.surface, &surfaceCapabilities)); err != nil {
		return errors.New("vk.GetPhysicalDeviceSurfaceCapabilities(): " + err.Error())
	}

	// ImageFormat
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

	{
		var supported vk.Bool32
		if err := vk.Error(vk.GetPhysicalDeviceSurfaceSupport(v.physicalDevice, 0, v.surface, &supported)); err != nil {
			return errors.New("vk.GetPhysicalDeviceSurfaceSupport(): " + err.Error())
		}

		if !supported.B() {
			return fmt.Errorf("vk.GetPhysicalDeviceSurfaceSupport(): surface is not supported")
		}
	}

	var swapchain vk.Swapchain
	scci := vk.SwapchainCreateInfo{
		SType:           vk.StructureTypeSwapchainCreateInfo,
		Surface:         v.surface,
		MinImageCount:   v.configuration.SwapchainSize,
		ImageFormat:     surfaceFormats[0].Format,
		ImageColorSpace: surfaceFormats[0].ColorSpace,
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
	v.imageFormat = surfaceFormats[0].Format

	var numImages uint32
	if err := vk.Error(vk.GetSwapchainImages(v.logicalDevice, v.swapchain, &numImages, nil)); err != nil {
		return errors.New("vk.GetSwapchainImages(num): " + err.Error())
	}

	v.swapchainImages = make([]vk.Image, numImages)
	if err := vk.Error(vk.GetSwapchainImages(v.logicalDevice, v.swapchain, &numImages, v.swapchainImages)); err != nil {
		return errors.New("vk.GetSwapchainImages(images): " + err.Error())
	}

	if err := v.createImageViews(); err != nil {
		return err
	}

	shaders, err := loadShaders(v.logicalDevice, v.configuration.ShaderDirectory)
	if err != nil {
		return err
	}

	shaderStages, err := createPipelineShaderStagesInfo(shaders)
	if err != nil {
		return err
	}

	/* Viewport and scissors creation */
	{
		v.viewport = vk.Viewport{
			X:        0,
			Y:        0,
			Width:    float32(v.configuration.ScreenWidth),
			Height:   float32(v.configuration.ScreenHeight),
			MinDepth: 0,
			MaxDepth: 1,
		}

		v.scissor = vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0,
				Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  v.configuration.ScreenWidth,
				Height: v.configuration.ScreenHeight,
			},
		}
	}

	// TODO: Depth and stencil testing VkPipelineDepthStencilStateCreateInfo
	// TODO: When making dynamic state changes refer to  VkPipelineDynamicStateCreateInfo
	// Dynamic state in vulkan-tutorial.com

	/* Pipeline Layout */
	{
		plci := vk.PipelineLayoutCreateInfo{
			SType: vk.StructureTypePipelineLayoutCreateInfo,
		}

		var pipelineLayout vk.PipelineLayout
		if err := vk.Error(vk.CreatePipelineLayout(v.logicalDevice, &plci, nil, &pipelineLayout)); err != nil {
			return errors.New("vk.CreatePipelineLayout(): " + err.Error())
		}
		v.pipelineLayout = pipelineLayout
	}

	/* Render pass */
	{
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
			AttachmentCount: 1,
			PAttachments:    []vk.AttachmentDescription{colorAttachment},
			SubpassCount:    1,
			PSubpasses:      []vk.SubpassDescription{subpass},
		}

		var renderPass vk.RenderPass
		if err := vk.Error(vk.CreateRenderPass(v.logicalDevice, &rpci, nil, &renderPass)); err != nil {
			return errors.New("vk.CreateRenderPass(): " + err.Error())
		}
		v.renderPass = renderPass
	}

	/* Pipeline */
	{
		pcbas := []vk.PipelineColorBlendAttachmentState{{
			ColorWriteMask:      0xF, // ColorComponentFlagBits -> R | G | B | A
			BlendEnable:         vk.False,
			SrcColorBlendFactor: vk.BlendFactorOne,
			DstColorBlendFactor: vk.BlendFactorZero,
			ColorBlendOp:        vk.BlendOpAdd,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpZero,
		}}

		gpci := []vk.GraphicsPipelineCreateInfo{{
			SType:      vk.StructureTypeGraphicsPipelineCreateInfo,
			StageCount: uint32(len(shaderStages)),
			PStages:    shaderStages,
			PVertexInputState: &vk.PipelineVertexInputStateCreateInfo{
				SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
			},
			PInputAssemblyState: &vk.PipelineInputAssemblyStateCreateInfo{
				SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
				Topology:               vk.PrimitiveTopologyTriangleList,
				PrimitiveRestartEnable: vk.False,
			},
			PViewportState: &vk.PipelineViewportStateCreateInfo{
				SType:         vk.StructureTypePipelineViewportStateCreateInfo,
				ViewportCount: 1,
				PViewports:    []vk.Viewport{v.viewport},
				ScissorCount:  1,
				PScissors:     []vk.Rect2D{v.scissor},
			},
			PRasterizationState: &vk.PipelineRasterizationStateCreateInfo{
				SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
				DepthClampEnable:        vk.False,
				RasterizerDiscardEnable: vk.False,
				PolygonMode:             vk.PolygonModeFill,
				LineWidth:               1.0,
				CullMode:                vk.CullModeFlags(vk.CullModeBackBit),
				FrontFace:               vk.FrontFaceClockwise,
			},
			PMultisampleState: &vk.PipelineMultisampleStateCreateInfo{
				SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
				RasterizationSamples: vk.SampleCount1Bit,
			},
			PDepthStencilState: nil,
			PColorBlendState: &vk.PipelineColorBlendStateCreateInfo{
				SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
				LogicOpEnable:   vk.False,
				LogicOp:         vk.LogicOpCopy,
				AttachmentCount: uint32(len(pcbas)),
				PAttachments:    pcbas,
				BlendConstants:  [4]float32{0.0, 0.0, 0.0, 0.0},
			},
			PDynamicState:      nil,
			Layout:             v.pipelineLayout,
			RenderPass:         v.renderPass,
			Subpass:            0,
			BasePipelineIndex:  -1,
			BasePipelineHandle: nil,
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
	}

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

func loadShaders(logicalDevice vk.Device, shaderDir string) ([]Shader, error) {
	var shaders []Shader
	shaderFiles, shaderTypes, err := loadShaderFilesFromDirectory(shaderDir)
	if err != nil {
		return nil, err
	}

	for idx, val := range shaderFiles {
		shader, err := NewVulkanShader(val, shaderTypes[idx], logicalDevice)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, shader)
	}
	return shaders, nil
}

func createPipelineShaderStagesInfo(shaders []Shader) ([]vk.PipelineShaderStageCreateInfo, error) {
	var pipelineShaderStagesInfo []vk.PipelineShaderStageCreateInfo
	for _, shader := range shaders {

		var stage vk.ShaderStageFlagBits
		switch shader.Type() {
		case VertexShaderType:
			stage = vk.ShaderStageFragmentBit
		case FragmentShaderType:
			stage = vk.ShaderStageVertexBit
		default:
			return nil, errors.New("unsupported shader type attempted creation")
		}

		var (
			innerShader vk.ShaderModule
			ok          bool
		)
		innerShader, ok = shader.Inner().(vk.ShaderModule)
		if !ok {
			return nil, errors.New("attempted shader cast is of invalid type")
		}

		pssci := vk.PipelineShaderStageCreateInfo{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  stage,
			Module: innerShader,
			PName:  shader.Name() + "\000",
		}

		pipelineShaderStagesInfo = append(pipelineShaderStagesInfo, pssci)
	}

	return pipelineShaderStagesInfo, nil
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

	for _, m := range v.shaders {
		vk.DestroyShaderModule(v.logicalDevice, m, nil)
	}

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
		shader:     shader,
		shaderType: shaderType,
		name:       shaderName,
	}, nil
}

// VulkanShader is a Vulkan specific shader
type VulkanShader struct {
	Shader

	name       string
	shaderType ShaderType
	shader     vk.ShaderModule
}

// Type implements interface
func (v VulkanShader) Type() ShaderType {
	return v.shaderType
}

// Inner implements interface
func (v VulkanShader) Inner() interface{} {
	return v.shader
}

// Name implements interface
func (v VulkanShader) Name() string {
	return v.name
}
