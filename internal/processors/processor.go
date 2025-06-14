package processors

import (
	"image"
)

// Processor defines the interface for image processors
type Processor interface {
	Process(img image.Image, params map[string]interface{}) (image.Image, error)
	ValidateParams(params map[string]interface{}) error
	Name() string
}