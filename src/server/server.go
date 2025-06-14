package server

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
	"github.com/mohammed-ysn/cluster-imager/pkg/middleware"
	"github.com/mohammed-ysn/cluster-imager/pkg/validation"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/crop"
	"github.com/mohammed-ysn/cluster-imager/src/image_processing/resize"
)

func StartServer() {
	// Initialize logger
	logger := logging.NewLogger(slog.LevelInfo)
	logger.Info("initializing server")

	// Create a new mux instead of using default
	mux := http.NewServeMux()
	mux.HandleFunc("/crop", cropHandler)
	mux.HandleFunc("/resize", resizeHandler)

	// Apply middleware
	handler := middleware.RequestLogging(logger)(mux)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB max header size
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		logger.Info("server started", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown or error
	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		// Give outstanding requests 5 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				logger.Error("forced shutdown failed", "error", err)
			}
		}
		logger.Info("server stopped")
	}
}

func processImage(w http.ResponseWriter, r *http.Request, processingFunc func(image.Image) (image.Image, error)) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get logger with request context
	logger := logging.NewLogger(slog.LevelInfo).WithContext(r.Context())
	logger.Debug("processing image request")

	// parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		logger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse uploaded data", http.StatusBadRequest)
		return
	}

	// get the uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		logger.Error("failed to get form file", "error", err)
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// decode the uploaded image
	inputImg, _, err := image.Decode(file)
	if err != nil {
		logger.Error("failed to decode image", "error", err)
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// call image processing function
	processedImage, err := processingFunc(inputImg)
	if err != nil {
		logger.Error("failed to process image", "error", err)
		http.Error(w, "Failed to process image", http.StatusBadRequest)
		return
	}

	// encode the processed image as JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, processedImage, nil)
	if err != nil {
		logger.Error("failed to encode image", "error", err)
		http.Error(w, "Failed to process image", http.StatusInternalServerError)
		return
	}

	// set the response headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))

	// write the processed image to the response writer
	_, err = w.Write(buf.Bytes())
	if err != nil {
		logger.Error("failed to write response", "error", err)
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
