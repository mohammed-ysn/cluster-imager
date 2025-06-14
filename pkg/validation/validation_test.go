package validation

import (
	"testing"
)

func TestValidateDimension(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		dimName string
		wantErr bool
	}{
		{"valid dimension", 100, "width", false},
		{"minimum dimension", 1, "width", false},
		{"maximum dimension", 10000, "width", false},
		{"negative dimension", -1, "width", true},
		{"zero dimension", 0, "width", true},
		{"exceeds maximum", 10001, "width", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDimension(tt.value, tt.dimName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDimension() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCropParams(t *testing.T) {
	tests := []struct {
		name                string
		x, y                int
		width, height       int
		imgWidth, imgHeight int
		wantErr             bool
	}{
		{"valid crop", 10, 10, 50, 50, 100, 100, false},
		{"crop at origin", 0, 0, 50, 50, 100, 100, false},
		{"crop full image", 0, 0, 100, 100, 100, 100, false},
		{"negative x", -10, 10, 50, 50, 100, 100, true},
		{"negative y", 10, -10, 50, 50, 100, 100, true},
		{"zero width", 10, 10, 0, 50, 100, 100, true},
		{"zero height", 10, 10, 50, 0, 100, 100, true},
		{"exceeds image width", 60, 10, 50, 50, 100, 100, true},
		{"exceeds image height", 10, 60, 50, 50, 100, 100, true},
		{"exactly at bounds", 50, 50, 50, 50, 100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCropParams(tt.x, tt.y, tt.width, tt.height, tt.imgWidth, tt.imgHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCropParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateResizeParams(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		wantErr bool
	}{
		{"valid resize", 100, 200, false},
		{"minimum size", 1, 1, false},
		{"maximum size", 10000, 10000, false},
		{"zero width", 0, 100, true},
		{"zero height", 100, 0, true},
		{"negative width", -100, 200, true},
		{"negative height", 100, -200, true},
		{"exceeds max width", 10001, 200, true},
		{"exceeds max height", 100, 10001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResizeParams(tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResizeParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
