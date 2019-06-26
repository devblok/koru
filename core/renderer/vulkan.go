package renderer

import (
	vk "github.com/vulkan-go/vulkan"
)

// Renderer describes the rendering machinery
type Renderer interface {
}

// NewVulkanRenderer creates a Vulkan API renderer
func NewVulkanRenderer(vkd vk.Device, vkpd vk.PhysicalDevice, vks vk.Surface, cfg Configuration) (Renderer, error) {

	/* Swapchain setup */
	var surfaceCapabilities vk.SurfaceCapabilities
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(vkpd, vks, &surfaceCapabilities)); err != nil {
		return nil, err
	}

	// ImageFormat
	var (
		surfaceFormatCount uint32
		surfaceFormats     []vk.SurfaceFormat
	)

	if err := vk.Error(vk.GetPhysicalDeviceSurfaceFormats(vkpd, vks, &surfaceFormatCount, nil)); err != nil {
		return nil, err
	}

	surfaceFormats = make([]vk.SurfaceFormat, surfaceFormatCount)
	if err := vk.Error(vk.GetPhysicalDeviceSurfaceFormats(vkpd, vks, &surfaceFormatCount, surfaceFormats)); err != nil {
		return nil, err
	}

	surfaceFormats[0].Deref()

	// PreTransform
	var preTransform vk.SurfaceTransformFlagBits
	requiredTransform := vk.SurfaceTransformIdentityBit
	if vk.SurfaceTransformFlagBits(surfaceCapabilities.SupportedTransforms) != 0 {
		preTransform = requiredTransform
	} else {
		preTransform = surfaceCapabilities.CurrentTransform
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
		SType:         vk.StructureTypeSwapchainCreateInfo,
		Surface:       vks,
		MinImageCount: cfg.SwapchainSize,
		ImageFormat:   surfaceFormats[0].Format,
		ImageExtent: vk.Extent2D{
			Width:  cfg.ScreenWidth,
			Height: cfg.ScreenHeight,
		},
		ImageUsage:     vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PreTransform:   preTransform,
		CompositeAlpha: compositeAlpha,
		PresentMode:    vk.PresentModeFifo,
		Clipped:        vk.True,
		OldSwapchain:   nil,
	}

	if err := vk.Error(vk.CreateSwapchain(vkd, &scci, nil, &swapchain)); err != nil {
		return nil, err
	}

	return &Vulkan{
		swapchain: swapchain,
	}, nil
}

// Vulkan is a Vulkan API renderer
type Vulkan struct {
	Renderer

	swapchain vk.Swapchain
}
