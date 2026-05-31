package processors

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
)

type CropProcessor struct{}

func NewCropProcessor() *CropProcessor {
	return &CropProcessor{}
}

func (p *CropProcessor) Process(img image.Image, params map[string]interface{}) (image.Image, error) {
	x, err := paramInt(params, "x")
	if err != nil {
		return nil, err
	}
	y, err := paramInt(params, "y")
	if err != nil {
		return nil, err
	}
	width, err := paramInt(params, "width")
	if err != nil {
		return nil, err
	}
	height, err := paramInt(params, "height")
	if err != nil {
		return nil, err
	}

	if err := validation.ValidateCropParams(x, y, width, height, img.Bounds().Dx(), img.Bounds().Dy()); err != nil {
		return nil, err
	}

	out := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(out, out.Bounds(), img, image.Pt(x, y), draw.Src)
	return out, nil
}

func (p *CropProcessor) ValidateParams(params map[string]interface{}) error {
	x, err := paramInt(params, "x")
	if err != nil {
		return err
	}
	y, err := paramInt(params, "y")
	if err != nil {
		return err
	}
	width, err := paramInt(params, "width")
	if err != nil {
		return err
	}
	height, err := paramInt(params, "height")
	if err != nil {
		return err
	}

	if x < 0 || y < 0 {
		return fmt.Errorf("coordinates cannot be negative")
	}
	if err := validation.ValidateDimension(width, "width"); err != nil {
		return err
	}
	return validation.ValidateDimension(height, "height")
}

func (p *CropProcessor) Name() string { return "crop" }
