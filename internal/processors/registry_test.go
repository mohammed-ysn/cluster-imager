package processors

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewRegistry()
		processor := NewCropProcessor()
		
		// Register processor
		err := registry.Register("test", processor)
		if err != nil {
			t.Fatalf("Failed to register processor: %v", err)
		}
		
		// Get processor
		retrieved, err := registry.Get("test")
		if err != nil {
			t.Fatalf("Failed to get processor: %v", err)
		}
		
		if retrieved != processor {
			t.Error("Retrieved processor does not match registered processor")
		}
	})
	
	t.Run("Get non-existent processor", func(t *testing.T) {
		registry := NewRegistry()
		
		_, err := registry.Get("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent processor")
		}
	})
	
	t.Run("Register duplicate", func(t *testing.T) {
		registry := NewRegistry()
		processor := NewCropProcessor()
		
		// Register first time
		err := registry.Register("test", processor)
		if err != nil {
			t.Fatalf("Failed to register processor: %v", err)
		}
		
		// Try to register again
		err = registry.Register("test", processor)
		if err == nil {
			t.Error("Expected error for duplicate registration")
		}
	})
	
	t.Run("List processors", func(t *testing.T) {
		registry := NewRegistry()
		
		// Register multiple processors
		registry.Register("crop", NewCropProcessor())
		registry.Register("resize", NewResizeProcessor())
		
		names := registry.List()
		if len(names) != 2 {
			t.Errorf("Expected 2 processors, got %d", len(names))
		}
		
		// Check that both names are present
		found := make(map[string]bool)
		for _, name := range names {
			found[name] = true
		}
		
		if !found["crop"] || !found["resize"] {
			t.Error("Missing expected processor names")
		}
	})
	
	t.Run("DefaultRegistry", func(t *testing.T) {
		registry := DefaultRegistry()
		
		// Check that default processors are registered
		names := registry.List()
		if len(names) < 2 {
			t.Error("DefaultRegistry should have at least 2 processors")
		}
		
		// Verify crop processor exists
		_, err := registry.Get("crop")
		if err != nil {
			t.Error("DefaultRegistry should have crop processor")
		}
		
		// Verify resize processor exists
		_, err = registry.Get("resize")
		if err != nil {
			t.Error("DefaultRegistry should have resize processor")
		}
	})
}