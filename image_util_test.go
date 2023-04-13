package main

import (
	"image"
	"image/png"
	"os"
	"testing"
)

// TODO: write more tests to cover edge cases

func TestLoadImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	imgFile, err := os.Create("test_load.png")
	if err != nil {
		t.Errorf("Error creating image file: %v", err)
	}
	defer os.Remove("test_load.png")
	err = png.Encode(imgFile, img)
	if err != nil {
		t.Errorf("Error encoding image: %v", err)
	}

	loadedImg, err := LoadImage("test_load.png")
	if err != nil {
		t.Errorf("Error loading image: %v", err)
	}
	if loadedImg.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.Errorf("Image dimensions are incorrect")
	}
}

func TestExportImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	defer os.Remove("test_output.png")
	err := ExportImage(img, "test_output.png", "png")
	if err != nil {
		t.Errorf("Error exporting image: %v", err)
	}
}
