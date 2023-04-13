package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
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

// ExportImage exports an Image object to a file on disk in the specified format.
// If the format is not specified, the format of the input image will be used.
func ExportImage(img image.Image, filePath string, format ...string) error {
	var outputFormat string
	if len(format) > 0 {
		outputFormat = format[0]
	} else {
		// determine the input image format by decoding it
		_, format, err := image.Decode(strings.NewReader(""))
		if err != nil {
			return err
		}
		outputFormat = format
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the image data in the specified format and write it to disk
	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported image format: %s", outputFormat)
	}
	if err != nil {
		return err
	}

	return nil
}
