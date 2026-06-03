package processors

import (
	"image"

	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
	"github.com/nfnt/resize"
)

type ResizeProcessor struct{}

func NewResizeProcessor() *ResizeProcessor {
	return &ResizeProcessor{}
}

func (p *ResizeProcessor) Process(img image.Image, params map[string]interface{}) (image.Image, error) {
	width, err := paramInt(params, "width")
	if err != nil {
		return nil, err
	}
	height, err := paramInt(params, "height")
	if err != nil {
		return nil, err
	}
	return resize.Resize(uint(width), uint(height), img, resize.Lanczos2), nil //nolint:gosec
}

func (p *ResizeProcessor) ValidateParams(params map[string]interface{}) error {
	width, err := paramInt(params, "width")
	if err != nil {
		return err
	}
	height, err := paramInt(params, "height")
	if err != nil {
		return err
	}
	return validation.ValidateResizeParams(width, height)
}

func (p *ResizeProcessor) Name() string { return "resize" }
