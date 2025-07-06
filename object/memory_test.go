package object

import (
	"testing"
)

func TestIntegerCaching(t *testing.T) {
	tests := []struct {
		value int64
		shouldBeCached bool
	}{
		{0, true},
		{1, true},
		{-1, true},
		{255, true},
		{-256, true},
		{256, false},  // Above cache range
		{-257, false}, // Below cache range
		{1000, false},
		{-1000, false},
	}

	for _, tt := range tests {
		obj1 := NewInteger(tt.value)
		obj2 := NewInteger(tt.value)
		
		if tt.shouldBeCached {
			if obj1 != obj2 {
				t.Errorf("Integer %d should be cached (same instance), but got different instances", tt.value)
			}
		} else {
			if obj1 == obj2 {
				t.Errorf("Integer %d should not be cached (different instances), but got same instance", tt.value)
			}
		}
		
		// Verify value is correct
		if obj1.Value != tt.value {
			t.Errorf("Integer value mismatch: expected %d, got %d", tt.value, obj1.Value)
		}
	}
}

func TestBooleanSingletons(t *testing.T) {
	// Test that TRUE instances are the same
	if TRUE != TRUE {
		t.Error("TRUE instances should be identical")
	}
	
	// Test that FALSE instances are the same  
	if FALSE != FALSE {
		t.Error("FALSE instances should be identical")
	}
	
	// Test that NULL instances are the same
	if NULL != NULL {
		t.Error("NULL instances should be identical")
	}
	
	// Test values are correct
	if TRUE.Value != true {
		t.Error("TRUE value should be true")
	}
	
	if FALSE.Value != false {
		t.Error("FALSE value should be false")
	}
}

func BenchmarkIntegerCreation(b *testing.B) {
	b.Run("Cached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This should use cached instances
			NewInteger(int64(i % 512 - 256))
		}
	})
	
	b.Run("Uncached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This should create new instances
			NewInteger(int64(i + 1000))
		}
	})
}

func BenchmarkBooleanAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = TRUE
		_ = FALSE
		_ = NULL
	}
}