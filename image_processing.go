package main

import (
	"image"
	"image/draw"
)

// cropImage crops the input image to the specified dimensions.
//
// The resulting image will have the specified width and height,
// and will be positioned at (x, y) within the input image.
func cropImage(inputImg image.Image, x, y, width, height int) image.Image {
	cropRect := image.Rect(0, 0, width, height)
	croppedImage := image.NewRGBA(cropRect)

	draw.Draw(croppedImage, cropRect, inputImg, image.Pt(x, y), draw.Src)

	return croppedImage
}
