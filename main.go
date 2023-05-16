package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mohammed-ysn/cluster-imager/image_processing"
)

func processImage(w http.ResponseWriter, r *http.Request, processingFunc func(image.Image) image.Image) {
	if r.Method != http.MethodPost {
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
	x, err := strconv.Atoi(r.URL.Query().Get("x"))
	if err != nil {
		http.Error(w, "Invalid value for 'x'", http.StatusBadRequest)
		return
	}

	y, err := strconv.Atoi(r.URL.Query().Get("y"))
	if err != nil {
		http.Error(w, "Invalid value for 'y'", http.StatusBadRequest)
		return
	}

	width, err := strconv.Atoi(r.URL.Query().Get("width"))
	if err != nil {
		http.Error(w, "Invalid value for 'width'", http.StatusBadRequest)
		return
	}

	height, err := strconv.Atoi(r.URL.Query().Get("height"))
	if err != nil {
		http.Error(w, "Invalid value for 'height'", http.StatusBadRequest)
		return
	}

	processImage(w, r, func(inputImg image.Image) image.Image {
		return image_processing.CropImage(inputImg, x, y, width, height)
	})
}

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	width, err := strconv.Atoi(r.URL.Query().Get("width"))
	if err != nil {
		http.Error(w, "Invalid value for 'width'", http.StatusBadRequest)
		return
	}

	height, err := strconv.Atoi(r.URL.Query().Get("height"))
	if err != nil {
		http.Error(w, "Invalid value for 'height'", http.StatusBadRequest)
		return
	}

	processImage(w, r, func(inputImg image.Image) image.Image {
		return image_processing.ResizeImage(inputImg, width, height)
	})
}

func main() {
	server := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB max header size
	}
	http.HandleFunc("/crop", cropHandler)
	http.HandleFunc("/resize", resizeHandler)
	fmt.Println("Server started on port 8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
