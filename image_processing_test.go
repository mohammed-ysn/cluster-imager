package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCropImage(t *testing.T) {
	// set up dimensions for the test image
	// and the desired cropped image
	beforeX, beforeY := 400, 400
	afterX, afterY := 100, 100
	cropStartX, cropStartY := 250, 250

	// create the test image
	testImg := image.NewRGBA(image.Rect(0, 0, beforeX, beforeY))
	draw.Draw(testImg, testImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(testImg, image.Rect(cropStartX, cropStartY, cropStartX+afterX, cropStartY+afterY), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	croppedImg := cropImage(testImg, cropStartX, cropStartY, afterX, afterY)

	// check cropped image dimensions
	if croppedImg.Bounds().Dx() != afterX || croppedImg.Bounds().Dy() != afterY {
		t.Errorf("Cropped image has unexpected dimensions: %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}

	// check all pixel colours in cropped image
	black_rgba := color.RGBAModel.Convert(color.Black)
	for x := 0; x < afterX; x++ {
		for y := 0; y < afterY; y++ {
			if croppedImg.At(x, y) != black_rgba {
				t.Errorf("Pixel at (%d, %d) is %v instead of %v", x, y, croppedImg.At(x, y), color.Black)
			}
		}
	}
}
