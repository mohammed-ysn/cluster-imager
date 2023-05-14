package image_processing

import (
	"image"
	"image/png"
	"os"
	"testing"
)

func TestLoadImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	imgFile, err := os.Create("test_load.png")
	if err != nil {
		t.Fatalf("Error creating image file: %v", err)
	}
	defer os.Remove("test_load.png")
	err = png.Encode(imgFile, img)
	if err != nil {
		t.Fatalf("Error encoding image: %v", err)
	}

	loadedImg, err := LoadImage("test_load.png")
	if err != nil {
		t.Fatalf("Error loading image: %v", err)
	}
	if loadedImg.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.Errorf("Image dimensions are incorrect")
	}
}

func TestLoadImageNonExistentFile(t *testing.T) {
	_, err := LoadImage("nonexistent_file.png")
	if err == nil {
		t.Fatalf("Expected error, but got none")
	}
}

func TestExportImagePNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	tmpfile := "test.png"
	defer os.Remove(tmpfile)
	if err := ExportImage(img, tmpfile); err != nil {
		t.Fatalf("ExportImage(%s) failed: %s", tmpfile, err)
	}

	if _, err := os.Stat(tmpfile); os.IsNotExist(err) {
		t.Errorf("%s was not created", tmpfile)
	}
}

func TestExportImageJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	tmpfile := "test.jpg"
	defer os.Remove(tmpfile)
	if err := ExportImage(img, tmpfile); err != nil {
		t.Fatalf("ExportImage(%s) failed: %s", tmpfile, err)
	}

	if _, err := os.Stat(tmpfile); os.IsNotExist(err) {
		t.Errorf("%s was not created", tmpfile)
	}
}

func TestExportImageUnsupportedFormat(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	unsupportedFile := "test.txt"
	defer os.Remove(unsupportedFile)
	if err := ExportImage(img, unsupportedFile); err == nil {
		t.Fatalf("ExportImage(%s) did not return an error", unsupportedFile)
	}
}
