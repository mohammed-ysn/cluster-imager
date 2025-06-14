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

func processImage(w http.ResponseWriter, r *http.Request, processingFunc func(image.Image) image.Image) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Handling new request")

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

	processImage(w, r, func(inputImg image.Image) image.Image {
		// Additional validation with actual image bounds
		if err := validation.ValidateCropParams(x, y, width, height, inputImg.Bounds().Dx(), inputImg.Bounds().Dy()); err != nil {
			// Return empty image on validation error - this will be handled better in next commit
			return image.NewRGBA(image.Rect(0, 0, 1, 1))
		}
		return crop.CropImage(inputImg, x, y, width, height)
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

	processImage(w, r, func(inputImg image.Image) image.Image {
		return resize.ResizeImage(inputImg, width, height)
	})
}
