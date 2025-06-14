package processors

import (
	"fmt"
	"image"

	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/crop"
)

// CropProcessor implements the Processor interface for cropping images
type CropProcessor struct{}

// NewCropProcessor creates a new crop processor
func NewCropProcessor() *CropProcessor {
	return &CropProcessor{}
}

// Process crops the image according to the provided parameters
func (p *CropProcessor) Process(img image.Image, params map[string]interface{}) (image.Image, error) {
	x, ok := params["x"].(int)
	if !ok {
		return nil, fmt.Errorf("x parameter is required and must be an integer")
	}
	
	y, ok := params["y"].(int)
	if !ok {
		return nil, fmt.Errorf("y parameter is required and must be an integer")
	}
	
	width, ok := params["width"].(int)
	if !ok {
		return nil, fmt.Errorf("width parameter is required and must be an integer")
	}
	
	height, ok := params["height"].(int)
	if !ok {
		return nil, fmt.Errorf("height parameter is required and must be an integer")
	}
	
	// Validate with actual image bounds
	if err := validation.ValidateCropParams(x, y, width, height, img.Bounds().Dx(), img.Bounds().Dy()); err != nil {
		return nil, err
	}
	
	return crop.CropImage(img, x, y, width, height), nil
}

// ValidateParams validates the crop parameters
func (p *CropProcessor) ValidateParams(params map[string]interface{}) error {
	x, ok := params["x"].(int)
	if !ok {
		return fmt.Errorf("x parameter is required and must be an integer")
	}
	
	y, ok := params["y"].(int)
	if !ok {
		return fmt.Errorf("y parameter is required and must be an integer")
	}
	
	width, ok := params["width"].(int)
	if !ok {
		return fmt.Errorf("width parameter is required and must be an integer")
	}
	
	height, ok := params["height"].(int)
	if !ok {
		return fmt.Errorf("height parameter is required and must be an integer")
	}
	
	// Basic validation without image bounds
	if x < 0 || y < 0 {
		return fmt.Errorf("coordinates cannot be negative")
	}
	
	if err := validation.ValidateDimension(width, "width"); err != nil {
		return err
	}
	
	if err := validation.ValidateDimension(height, "height"); err != nil {
		return err
	}
	
	return nil
}

// Name returns the processor name
func (p *CropProcessor) Name() string {
	return "crop"
}