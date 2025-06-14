package server

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/crop"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/resize"
)

func StartServer() {
	server := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB max header size
	}

	// register the routes
	http.HandleFunc("/crop", cropHandler)
	http.HandleFunc("/resize", resizeHandler)

	// start the server
	log.Println("Server started on port 8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func processImage(w http.ResponseWriter, r *http.Request, processingFunc func(image.Image) (image.Image, error)) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Handling new request")

	// parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Failed to parse uploaded data", http.StatusBadRequest)
		return
	}

	// get the uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// decode the uploaded image
	inputImg, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// call image processing function
	processedImage, err := processingFunc(inputImg)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		http.Error(w, "Failed to process image", http.StatusBadRequest)
		return
	}

	// encode the processed image as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, processedImage, nil)
	if err != nil {
		log.Printf("Error encoding image: %v", err)
		http.Error(w, "Failed to process image", http.StatusInternalServerError)
		return
	}

	// set the response headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	// write the processed image to the response writer
	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing response: %v", err)
		// Can't send error response after headers are written
		return
	}
}

func cropHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	x, err := strconv.Atoi(params.Get("x"))
	if err != nil {
		http.Error(w, "Invalid value for 'x'", http.StatusBadRequest)
		return
	}

	y, err := strconv.Atoi(params.Get("y"))
	if err != nil {
		http.Error(w, "Invalid value for 'y'", http.StatusBadRequest)
		return
	}

	width, err := strconv.Atoi(params.Get("width"))
	if err != nil {
		http.Error(w, "Invalid value for 'width'", http.StatusBadRequest)
		return
	}

	height, err := strconv.Atoi(params.Get("height"))
	if err != nil {
		http.Error(w, "Invalid value for 'height'", http.StatusBadRequest)
		return
	}

	// Validate parameters before processing
	if x < 0 || y < 0 {
		http.Error(w, "Coordinates cannot be negative", http.StatusBadRequest)
		return
	}
	if err := validation.ValidateDimension(width, "width"); err != nil {
		http.Error(w, "Invalid width parameter", http.StatusBadRequest)
		return
	}
	if err := validation.ValidateDimension(height, "height"); err != nil {
		http.Error(w, "Invalid height parameter", http.StatusBadRequest)
		return
	}

	processImage(w, r, func(inputImg image.Image) (image.Image, error) {
		// Additional validation with actual image bounds
		if err := validation.ValidateCropParams(x, y, width, height, inputImg.Bounds().Dx(), inputImg.Bounds().Dy()); err != nil {
			return nil, err
		}
		return crop.CropImage(inputImg, x, y, width, height), nil
	})
}

func resizeHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	width, err := strconv.Atoi(params.Get("width"))
	if err != nil {
		http.Error(w, "Invalid value for 'width'", http.StatusBadRequest)
		return
	}

	height, err := strconv.Atoi(params.Get("height"))
	if err != nil {
		http.Error(w, "Invalid value for 'height'", http.StatusBadRequest)
		return
	}

	// Validate resize parameters
	if err := validation.ValidateResizeParams(width, height); err != nil {
		http.Error(w, "Invalid resize parameters", http.StatusBadRequest)
		return
	}

	processImage(w, r, func(inputImg image.Image) (image.Image, error) {
		return resize.ResizeImage(inputImg, width, height), nil
	})
}
