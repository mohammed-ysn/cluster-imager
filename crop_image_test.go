package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCropImageDimensionsWithZeroPosition(t *testing.T) {
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))

	croppedImg := cropImage(testImg, 0, 0, 50, 50)

	// check cropped image dimensions
	if croppedImg.Bounds().Dx() != 50 || croppedImg.Bounds().Dy() != 50 {
		t.Errorf("Expected dimensions: 50x50. Actual dimensions: %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}
}

func TestCropImageWithNonZeroPosition(t *testing.T) {
	// set up dimensions for the test image
	// and the desired cropped image
	beforeX, beforeY := 400, 400
	afterX, afterY := 10, 10
	cropStartX, cropStartY := 250, 250

	// create the test image
	testImg := image.NewRGBA(image.Rect(0, 0, beforeX, beforeY))
	draw.Draw(testImg, testImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(testImg, image.Rect(cropStartX, cropStartY, cropStartX+afterX, cropStartY+afterY), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	croppedImg := cropImage(testImg, cropStartX, cropStartY, afterX, afterY)

	if croppedImg.Bounds().Dx() != afterX || croppedImg.Bounds().Dy() != afterY {
		t.Errorf("Cropped image has unexpected dimensions: %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}

	// check all pixel colours in cropped image
	for x := 0; x < afterX; x++ {
		for y := 0; y < afterY; y++ {
			r, g, b, _ := croppedImg.At(x, y).RGBA()
			if r != 0 || g != 0 || b != 0 {
				t.Errorf("Pixel at (%d, %d) is (%v,%v,%v) instead of (0,0,0)", x, y, r, g, b)
			}
		}
	}
}

func TestCropImageWithWidthGreaterThanImageWidth(t *testing.T) {
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))

	croppedImg := cropImage(testImg, 0, 0, 150, 50)

	if croppedImg.Bounds().Dx() != 100 || croppedImg.Bounds().Dy() != 50 {
		t.Errorf("Expected dimensions: 100x50. Actual dimensions: %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}
}

func TestCropImageWithHeightGreaterThanImageHeight(t *testing.T) {
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))

	croppedImg := cropImage(testImg, 0, 0, 50, 150)

	if croppedImg.Bounds().Dx() != 50 || croppedImg.Bounds().Dy() != 100 {
		t.Errorf("Expected dimensions: 50x100. Actual dimensions: %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}
}
