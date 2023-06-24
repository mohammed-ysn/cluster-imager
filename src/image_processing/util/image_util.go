package util

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// LoadImage loads an image file from disk and returns an Image object.
func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func ExportImage(img image.Image, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var format string
	if ext := filepath.Ext(filePath); ext == ".jpg" || ext == ".jpeg" {
		format = "jpeg"
	} else if ext == ".png" {
		format = "png"
	} else {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	switch format {
	case "jpeg":
		if err := jpeg.Encode(file, img, nil); err != nil {
			return err
		}
	case "png":
		if err := png.Encode(file, img); err != nil {
			return err
		}
	}

	return nil
}
