package processors

import (
	"fmt"
	"image"
)

// Processor defines the interface for image processors
type Processor interface {
	Process(img image.Image, params map[string]interface{}) (image.Image, error)
	ValidateParams(params map[string]interface{}) error
	Name() string
}

// toInt extracts an int from a param value that may be int or float64 (JSON round-trip).
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	}
	return 0, false
}

func paramInt(params map[string]any, key string) (int, error) {
	v, ok := toInt(params[key])
	if !ok {
		return 0, fmt.Errorf("%s parameter is required and must be an integer", key)
	}
	return v, nil
}
