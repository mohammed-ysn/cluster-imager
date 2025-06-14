package validation

import (
	"errors"
	"fmt"
)

const (
	MaxImageDimension = 10000
	MinImageDimension = 1
	MaxFileSize       = 10 << 20 // 10MB
)

var (
	ErrInvalidDimension = errors.New("invalid dimension")
	ErrDimensionTooLarge = errors.New("dimension exceeds maximum allowed size")
	ErrDimensionTooSmall = errors.New("dimension below minimum allowed size")
	ErrNegativeDimension = errors.New("dimension cannot be negative")
)

// ValidateDimension validates a single dimension value
func ValidateDimension(value int, name string) error {
	if value < 0 {
		return fmt.Errorf("%s: %w", name, ErrNegativeDimension)
	}
	if value < MinImageDimension {
		return fmt.Errorf("%s must be at least %d: %w", name, MinImageDimension, ErrDimensionTooSmall)
	}
	if value > MaxImageDimension {
		return fmt.Errorf("%s cannot exceed %d: %w", name, MaxImageDimension, ErrDimensionTooLarge)
	}
	return nil
}

// ValidateCropParams validates cropping parameters
func ValidateCropParams(x, y, width, height int, imgWidth, imgHeight int) error {
	// Validate individual dimensions
	if err := ValidateDimension(width, "width"); err != nil {
		return err
	}
	if err := ValidateDimension(height, "height"); err != nil {
		return err
	}
	
	// X and Y can be 0, but not negative
	if x < 0 {
		return fmt.Errorf("x coordinate cannot be negative")
	}
	if y < 0 {
		return fmt.Errorf("y coordinate cannot be negative")
	}
	
	// Check if crop area is within image bounds
	if x+width > imgWidth {
		return fmt.Errorf("crop area exceeds image width")
	}
	if y+height > imgHeight {
		return fmt.Errorf("crop area exceeds image height")
	}
	
	return nil
}

// ValidateResizeParams validates resize parameters
func ValidateResizeParams(width, height int) error {
	if err := ValidateDimension(width, "width"); err != nil {
		return err
	}
	if err := ValidateDimension(height, "height"); err != nil {
		return err
	}
	return nil
}