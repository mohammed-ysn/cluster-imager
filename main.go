package main

import (
	"bytes"
	"fmt"
	"github.com/mohammed-ysn/cluster-imager/image_processing"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
)

func processImage(w http.ResponseWriter, r *http.Request, processingFunc func(image.Image) image.Image) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Handling new request")

	// parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get the uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// decode the uploaded image
	inputImg, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// call image processing function
	processedImage := processingFunc(inputImg)

	// encode the processed image as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, processedImage, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set the response headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	// write the processed image to the response writer
	_, err = w.Write(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func cropHandler(w http.ResponseWriter, r *http.Request) {
	processImage(w, r, func(inputImg image.Image) image.Image {
		// TODO: remove hard-coded values
		return image_processing.CropImage(inputImg, 100, 50, 60, 60)
	})
}

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	processImage(w, r, func(inputImg image.Image) image.Image {
		// TODO: remove hard-coded values
		return image_processing.ResizeImage(inputImg, 200, 150)
	})
}

func main() {
	http.HandleFunc("/crop", cropHandler)
	http.HandleFunc("/resize", resizeHandler)
	fmt.Println("Server started on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
