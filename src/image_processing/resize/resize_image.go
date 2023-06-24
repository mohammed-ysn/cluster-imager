package resize

import (
	"github.com/nfnt/resize"
	"image"
)

// ResizeImage resizes an input image to the specified width and height.
//
// It takes in an image object as input, along with the desired width and height of the output image.
// It returns the resized image object.
func ResizeImage(inputImg image.Image, width, height int) image.Image {
	resizedImg := resize.Resize(uint(width), uint(height), inputImg, resize.Lanczos2)
	return resizedImg
}
