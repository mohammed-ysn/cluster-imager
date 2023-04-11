package main

import (
	"image"
	"testing"
)

func TestResizeImage(t *testing.T) {
	testImg := image.NewRGBA(image.Rect(0, 0, 100, 100))

	resizedImg := resizeImage(testImg, 50, 50)

	if resizedImg.Bounds().Dx() != 50 || resizedImg.Bounds().Dy() != 50 {
		t.Errorf("Resized image has unexpected dimensions: %dx%d", resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy())
	}
}
