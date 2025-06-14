package processors

import (
	"image"
	"testing"
)

func TestCropProcessor_Name(t *testing.T) {
	p := NewCropProcessor()
	if p.Name() != "crop" {
		t.Errorf("Expected name 'crop', got '%s'", p.Name())
	}
}

func TestCropProcessor_ValidateParams(t *testing.T) {
	p := NewCropProcessor()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"x":      10,
				"y":      20,
				"width":  100,
				"height": 200,
			},
			wantErr: false,
		},
		{
			name: "negative x",
			params: map[string]interface{}{
				"x":      -10,
				"y":      20,
				"width":  100,
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "negative y",
			params: map[string]interface{}{
				"x":      10,
				"y":      -20,
				"width":  100,
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "zero width",
			params: map[string]interface{}{
				"x":      10,
				"y":      20,
				"width":  0,
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "missing x",
			params: map[string]interface{}{
				"y":      20,
				"width":  100,
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "wrong type for width",
			params: map[string]interface{}{
				"x":      10,
				"y":      20,
				"width":  "100",
				"height": 200,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCropProcessor_Process(t *testing.T) {
	p := NewCropProcessor()

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid crop within bounds",
			params: map[string]interface{}{
				"x":      10,
				"y":      10,
				"width":  50,
				"height": 50,
			},
			wantErr: false,
		},
		{
			name: "crop exceeds image bounds",
			params: map[string]interface{}{
				"x":      150,
				"y":      150,
				"width":  100,
				"height": 100,
			},
			wantErr: true,
		},
		{
			name: "invalid params",
			params: map[string]interface{}{
				"x":      10,
				"y":      10,
				"width":  "50",
				"height": 50,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Process(img, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("Process() returned nil result without error")
			}
		})
	}
}
