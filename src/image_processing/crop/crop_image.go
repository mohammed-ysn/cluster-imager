package crop

import (
	"image"
	"image/draw"
)

// CropImage crops the input image to the specified dimensions.
//
// The resulting image will have the specified width and height,
// starting from position (x, y) within the input image.
func CropImage(inputImg image.Image, x, y, width, height int) image.Image {
	// Create a new image with the cropped dimensions
	croppedImage := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Draw the cropped portion onto the new image
	draw.Draw(croppedImage, croppedImage.Bounds(), inputImg, image.Pt(x, y), draw.Src)
	
	return croppedImage
}
