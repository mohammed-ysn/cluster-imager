package processors

import (
	"image"
	"testing"
)

func TestResizeProcessor_Name(t *testing.T) {
	p := NewResizeProcessor()
	if p.Name() != "resize" {
		t.Errorf("Expected name 'resize', got '%s'", p.Name())
	}
}

func TestResizeProcessor_ValidateParams(t *testing.T) {
	p := NewResizeProcessor()
	
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"width":  100,
				"height": 200,
			},
			wantErr: false,
		},
		{
			name: "zero width",
			params: map[string]interface{}{
				"width":  0,
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "negative height",
			params: map[string]interface{}{
				"width":  100,
				"height": -200,
			},
			wantErr: true,
		},
		{
			name: "missing width",
			params: map[string]interface{}{
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "wrong type for width",
			params: map[string]interface{}{
				"width":  "100",
				"height": 200,
			},
			wantErr: true,
		},
		{
			name: "exceeds max dimension",
			params: map[string]interface{}{
				"width":  20000,
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

func TestResizeProcessor_Process(t *testing.T) {
	p := NewResizeProcessor()
	
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	
	tests := []struct {
		name         string
		params       map[string]interface{}
		wantErr      bool
		expectWidth  int
		expectHeight int
	}{
		{
			name: "valid resize",
			params: map[string]interface{}{
				"width":  100,
				"height": 100,
			},
			wantErr:      false,
			expectWidth:  100,
			expectHeight: 100,
		},
		{
			name: "invalid params",
			params: map[string]interface{}{
				"width":  "100",
				"height": 100,
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
			if !tt.wantErr {
				if result == nil {
					t.Error("Process() returned nil result without error")
				} else {
					bounds := result.Bounds()
					if bounds.Dx() != tt.expectWidth || bounds.Dy() != tt.expectHeight {
						t.Errorf("Process() resulted in wrong dimensions: got %dx%d, want %dx%d",
							bounds.Dx(), bounds.Dy(), tt.expectWidth, tt.expectHeight)
					}
				}
			}
		})
	}
}