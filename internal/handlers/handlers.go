package handlers

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"

	"github.com/mohammed-ysn/cluster-imager/internal/processors"
	"github.com/mohammed-ysn/cluster-imager/pkg/logging"
)

// Handlers contains HTTP handlers
type Handlers struct {
	logger   *logging.Logger
	registry *processors.Registry
}

// New creates new handlers instance
func New(logger *logging.Logger, registry *processors.Registry) *Handlers {
	return &Handlers{
		logger:   logger,
		registry: registry,
	}
}

// CropHandler handles image cropping requests
func (h *Handlers) CropHandler(w http.ResponseWriter, r *http.Request) {
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
	
	processorParams := map[string]interface{}{
		"x":      x,
		"y":      y,
		"width":  width,
		"height": height,
	}
	
	h.processImage(w, r, "crop", processorParams)
}

// ResizeHandler handles image resizing requests
func (h *Handlers) ResizeHandler(w http.ResponseWriter, r *http.Request) {
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
	
	processorParams := map[string]interface{}{
		"width":  width,
		"height": height,
	}
	
	h.processImage(w, r, "resize", processorParams)
}

// processImage is a generic image processing handler
func (h *Handlers) processImage(w http.ResponseWriter, r *http.Request, processorName string, params map[string]interface{}) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get logger with request context
	logger := h.logger.WithContext(r.Context())
	logger.Debug("processing image request", "processor", processorName)
	
	// Get processor
	processor, err := h.registry.Get(processorName)
	if err != nil {
		logger.Error("processor not found", "processor", processorName, "error", err)
		http.Error(w, "Invalid processor", http.StatusInternalServerError)
		return
	}
	
	// Validate parameters
	if err := processor.ValidateParams(params); err != nil {
		logger.Debug("invalid parameters", "error", err)
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		logger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse uploaded data", http.StatusBadRequest)
		return
	}
	
	// Get uploaded file
	file, _, err := r.FormFile("image")
	if err != nil {
		logger.Error("failed to get form file", "error", err)
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Decode image
	inputImg, _, err := image.Decode(file)
	if err != nil {
		logger.Error("failed to decode image", "error", err)
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}
	
	// Process image
	processedImage, err := processor.Process(inputImg, params)
	if err != nil {
		logger.Error("failed to process image", "processor", processorName, "error", err)
		http.Error(w, "Failed to process image", http.StatusBadRequest)
		return
	}
	
	// Encode result
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, processedImage, nil)
	if err != nil {
		logger.Error("failed to encode image", "error", err)
		http.Error(w, "Failed to process image", http.StatusInternalServerError)
		return
	}
	
	// Send response
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	
	_, err = w.Write(buf.Bytes())
	if err != nil {
		logger.Error("failed to write response", "error", err)
	}
}