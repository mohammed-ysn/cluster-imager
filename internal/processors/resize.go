package processors

import (
	"fmt"
	"image"

	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/resize"
)

// ResizeProcessor implements the Processor interface for resizing images
type ResizeProcessor struct{}

// NewResizeProcessor creates a new resize processor
func NewResizeProcessor() *ResizeProcessor {
	return &ResizeProcessor{}
}

// Process resizes the image according to the provided parameters
func (p *ResizeProcessor) Process(img image.Image, params map[string]interface{}) (image.Image, error) {
	width, ok := params["width"].(int)
	if !ok {
		return nil, fmt.Errorf("width parameter is required and must be an integer")
	}

	height, ok := params["height"].(int)
	if !ok {
		return nil, fmt.Errorf("height parameter is required and must be an integer")
	}

	return resize.ResizeImage(img, width, height), nil
}

// ValidateParams validates the resize parameters
func (p *ResizeProcessor) ValidateParams(params map[string]interface{}) error {
	width, ok := params["width"].(int)
	if !ok {
		return fmt.Errorf("width parameter is required and must be an integer")
	}

	height, ok := params["height"].(int)
	if !ok {
		return fmt.Errorf("height parameter is required and must be an integer")
	}

	return validation.ValidateResizeParams(width, height)
}

// Name returns the processor name
func (p *ResizeProcessor) Name() string {
	return "resize"
}
