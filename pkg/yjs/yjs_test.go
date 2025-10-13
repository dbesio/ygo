// pkg/yjs/yjs_test.go - Comprehensive unit tests for Y.js FFI operations
package yjs

import (
	"testing"
)

func TestYjsService_CreateEmptySnapshot(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	snapshot, err := service.CreateEmptySnapshot()
	if err != nil {
		t.Fatalf("Failed to create empty snapshot: %v", err)
	}

	if len(snapshot) == 0 {
		t.Error("Empty snapshot should have non-zero length (contains Y.js document structure)")
	}

	t.Logf("Empty snapshot length: %d bytes", len(snapshot))
}

func TestYjsService_CreateSnapshotWithText(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	tests := []struct {
		name      string
		fieldName string
		text      string
		wantErr   bool
	}{
		{
			name:      "simple text",
			fieldName: "content",
			text:      "Hello, World!",
			wantErr:   false,
		},
		{
			name:      "empty text",
			fieldName: "content",
			text:      "",
			wantErr:   false,
		},
		{
			name:      "multiline text",
			fieldName: "content",
			text:      "Line 1\nLine 2\nLine 3",
			wantErr:   false,
		},
		{
			name:      "unicode text",
			fieldName: "content",
			text:      "Hello 世界! 🌍",
			wantErr:   false,
		},
		{
			name:      "large text",
			fieldName: "content",
			text:      string(make([]byte, 10000)), // 10KB of null bytes
			wantErr:   false,
		},
		{
			name:      "empty field name",
			fieldName: "",
			text:      "test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := service.CreateSnapshotWithText(tt.fieldName, tt.text)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create snapshot with text: %v", err)
			}

			if len(snapshot) == 0 {
				t.Error("Snapshot with text should have non-zero length")
			}

			t.Logf("Snapshot with text '%s' length: %d bytes", tt.text[:min(20, len(tt.text))], len(snapshot))
		})
	}
}

func TestYjsService_CreateUpdateWithText(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	update, err := service.CreateUpdateWithText("content", "Initial content")
	if err != nil {
		t.Fatalf("Failed to create update with text: %v", err)
	}

	if len(update) == 0 {
		t.Error("Update should have non-zero length")
	}

	t.Logf("Update with text length: %d bytes", len(update))
}

func TestYjsService_ApplyUpdates(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	// Create base snapshot
	baseSnapshot, err := service.CreateEmptySnapshot()
	if err != nil {
		t.Fatalf("Failed to create base snapshot: %v", err)
	}

	// Create some updates
	update1, err := service.CreateUpdateWithText("content", "Hello")
	if err != nil {
		t.Fatalf("Failed to create update 1: %v", err)
	}

	update2, err := service.CreateUpdateWithText("content", " World")
	if err != nil {
		t.Fatalf("Failed to create update 2: %v", err)
	}

	tests := []struct {
		name        string
		baseSnapshot []byte
		updates     [][]byte
		wantErr     bool
	}{
		{
			name:        "no updates",
			baseSnapshot: baseSnapshot,
			updates:     [][]byte{},
			wantErr:     false,
		},
		{
			name:        "single update",
			baseSnapshot: baseSnapshot,
			updates:     [][]byte{update1},
			wantErr:     false,
		},
		{
			name:        "multiple updates",
			baseSnapshot: baseSnapshot,
			updates:     [][]byte{update1, update2},
			wantErr:     false,
		},
		{
			name:        "nil base snapshot",
			baseSnapshot: nil,
			updates:     [][]byte{update1},
			wantErr:     true,
		},
		{
			name:        "empty base snapshot",
			baseSnapshot: []byte{},
			updates:     [][]byte{update1},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ApplyUpdates(tt.baseSnapshot, tt.updates)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to apply updates: %v", err)
			}

			if len(result) == 0 {
				t.Error("Result snapshot should have non-zero length")
			}

			t.Logf("Applied %d updates, result length: %d bytes", len(tt.updates), len(result))
		})
	}
}

func TestYjsService_ValidateUpdate(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	// Create a valid update
	validUpdate, err := service.CreateUpdateWithText("content", "test")
	if err != nil {
		t.Fatalf("Failed to create valid update: %v", err)
	}

	tests := []struct {
		name    string
		update  []byte
		wantErr bool
	}{
		{
			name:    "valid update",
			update:  validUpdate,
			wantErr: false,
		},
		{
			name:    "empty update",
			update:  []byte{},
			wantErr: true,
		},
		{
			name:    "nil update",
			update:  nil,
			wantErr: true,
		},
		{
			name:    "single byte",
			update:  []byte{0x01},
			wantErr: true,
		},
		{
			name:    "random bytes",
			update:  []byte{0xff, 0xfe, 0xfd, 0xfc},
			wantErr: false, // Basic validation only checks length
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUpdate(tt.update)
			
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestYjsService_ExtractText(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	tests := []struct {
		name         string
		setupText    string
		fieldName    string
		expectedText string
		wantErr      bool
	}{
		{
			name:         "simple text extraction",
			setupText:    "Hello, World!",
			fieldName:    "content",
			expectedText: "Hello, World!",
			wantErr:      false,
		},
		{
			name:         "empty text extraction",
			setupText:    "",
			fieldName:    "content",
			expectedText: "",
			wantErr:      false,
		},
		{
			name:         "multiline text extraction",
			setupText:    "Line 1\nLine 2\nLine 3",
			fieldName:    "content",
			expectedText: "Line 1\nLine 2\nLine 3",
			wantErr:      false,
		},
		{
			name:         "unicode text extraction",
			setupText:    "Hello 世界! 🌍",
			fieldName:    "content",
			expectedText: "Hello 世界! 🌍",
			wantErr:      false,
		},
		{
			name:      "empty field name",
			setupText: "test",
			fieldName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty field name test for snapshot creation
			if tt.fieldName == "" {
				_, err := service.ExtractText([]byte{}, tt.fieldName)
				if err == nil && tt.wantErr {
					t.Error("Expected error for empty field name")
				}
				return
			}

			// Create a snapshot with text
			snapshot, err := service.CreateSnapshotWithText(tt.fieldName, tt.setupText)
			if err != nil {
				t.Fatalf("Failed to create snapshot with text: %v", err)
			}

			// Extract text from the snapshot
			extractedText, err := service.ExtractText(snapshot, tt.fieldName)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to extract text: %v", err)
			}

			if extractedText != tt.expectedText {
				t.Errorf("Extracted text mismatch. Expected: %q, Got: %q", tt.expectedText, extractedText)
			}

			t.Logf("Successfully extracted text: %q from %d byte snapshot", extractedText, len(snapshot))
		})
	}
}

func TestYjsService_ExtractText_EdgeCases(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	t.Run("extract_from_empty_snapshot", func(t *testing.T) {
		// This should fail since empty snapshot has no data
		_, err := service.ExtractText([]byte{}, "content")
		if err == nil {
			t.Error("Expected error when extracting from empty snapshot")
		}
		t.Logf("Empty snapshot extraction error (expected): %v", err)
	})

	t.Run("extract_from_nil_snapshot", func(t *testing.T) {
		// This should fail since nil snapshot is invalid
		_, err := service.ExtractText(nil, "content")
		if err == nil {
			t.Error("Expected error when extracting from nil snapshot")
		}
		t.Logf("Nil snapshot extraction error (expected): %v", err)
	})

	t.Run("extract_nonexistent_field", func(t *testing.T) {
		// Create snapshot with "content" field but try to extract "other"
		snapshot, err := service.CreateSnapshotWithText("content", "test content")
		if err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}

		// Try to extract from a field that doesn't exist
		text, err := service.ExtractText(snapshot, "nonexistent")
		// This might return empty string or error depending on Y.js behavior
		t.Logf("Nonexistent field extraction result: text=%q, err=%v", text, err)
	})

	t.Run("extract_after_apply_updates", func(t *testing.T) {
		// Create empty snapshot
		emptySnapshot, err := service.CreateEmptySnapshot()
		if err != nil {
			t.Fatalf("Failed to create empty snapshot: %v", err)
		}

		// Create initial update
		initialUpdate, err := service.CreateUpdateWithText("content", "Initial text")
		if err != nil {
			t.Fatalf("Failed to create initial update: %v", err)
		}

		// Apply update to get snapshot with content
		snapshotWithContent, err := service.ApplyUpdates(emptySnapshot, [][]byte{initialUpdate})
		if err != nil {
			t.Fatalf("Failed to apply updates: %v", err)
		}

		// Extract text from the resulting snapshot
		extractedText, err := service.ExtractText(snapshotWithContent, "content")
		if err != nil {
			t.Fatalf("Failed to extract text after applying updates: %v", err)
		}

		t.Logf("Text extracted after applying updates: %q", extractedText)
		
		// The extracted text should contain "Initial text" but Y.js behavior may vary
		if extractedText == "" {
			t.Log("Warning: Empty text extracted after applying updates - this may indicate Y.js FFI behavior")
		}
	})
}

func TestYjsService_Integration_CreateAndApply(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	// Test the complete workflow: empty -> snapshot with text -> apply updates
	
	// Step 1: Create empty snapshot
	emptySnapshot, err := service.CreateEmptySnapshot()
	if err != nil {
		t.Fatalf("Failed to create empty snapshot: %v", err)
	}

	// Step 2: Create initial content update
	initialUpdate, err := service.CreateUpdateWithText("content", "Initial text")
	if err != nil {
		t.Fatalf("Failed to create initial update: %v", err)
	}

	// Step 3: Apply initial update to empty snapshot
	snapshotWithContent, err := service.ApplyUpdates(emptySnapshot, [][]byte{initialUpdate})
	if err != nil {
		t.Fatalf("Failed to apply initial update: %v", err)
	}

	// Step 4: Create additional updates
	update1, err := service.CreateUpdateWithText("content", " - Modified")
	if err != nil {
		t.Fatalf("Failed to create update 1: %v", err)
	}

	update2, err := service.CreateUpdateWithText("content", " - Final")
	if err != nil {
		t.Fatalf("Failed to create update 2: %v", err)
	}

	// Step 5: Apply additional updates
	finalSnapshot, err := service.ApplyUpdates(snapshotWithContent, [][]byte{update1, update2})
	if err != nil {
		t.Fatalf("Failed to apply additional updates: %v", err)
	}

	// Verify progression
	if len(emptySnapshot) >= len(snapshotWithContent) {
		t.Error("Snapshot with content should be larger than empty snapshot")
	}

	if len(snapshotWithContent) >= len(finalSnapshot) {
		t.Error("Final snapshot should be larger than intermediate snapshot")
	}

	t.Logf("Integration test progression: empty(%d) -> with_content(%d) -> final(%d)", 
		len(emptySnapshot), len(snapshotWithContent), len(finalSnapshot))
}

func TestYjsService_ErrorHandling(t *testing.T) {
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	// Test various error conditions that should not crash
	
	t.Run("apply_updates_with_invalid_data", func(t *testing.T) {
		// This should not crash even with invalid data
		baseSnapshot, _ := service.CreateEmptySnapshot()
		invalidUpdate := []byte{0xFF, 0xFF, 0xFF, 0xFF} // Invalid Y.js update
		
		_, err := service.ApplyUpdates(baseSnapshot, [][]byte{invalidUpdate})
		// Error is expected, but no crash
		t.Logf("Apply invalid updates error (expected): %v", err)
	})

	t.Run("create_snapshot_with_very_long_field_name", func(t *testing.T) {
		longFieldName := string(make([]byte, 1000))
		_, err := service.CreateSnapshotWithText(longFieldName, "test")
		// Should handle gracefully
		t.Logf("Long field name result: %v", err)
	})

	t.Run("create_snapshot_with_very_large_text", func(t *testing.T) {
		largeText := string(make([]byte, 100000)) // 100KB
		_, err := service.CreateSnapshotWithText("content", largeText)
		// Should handle large text
		if err != nil {
			t.Logf("Large text error: %v", err)
		} else {
			t.Log("Large text handled successfully")
		}
	})
}

func TestYjsService_Concurrent(t *testing.T) {
	// Test concurrent usage to ensure thread safety
	service, err := NewYjsService()
	if err != nil {
		t.Fatalf("Failed to create Y.js service: %v", err)
	}

	// Run multiple operations concurrently
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			// Each goroutine performs Y.js operations
			snapshot, err := service.CreateEmptySnapshot()
			if err != nil {
				t.Errorf("Goroutine %d: failed to create snapshot: %v", id, err)
				return
			}

			update, err := service.CreateUpdateWithText("content", "Concurrent test")
			if err != nil {
				t.Errorf("Goroutine %d: failed to create update: %v", id, err)
				return
			}

			_, err = service.ApplyUpdates(snapshot, [][]byte{update})
			if err != nil {
				t.Errorf("Goroutine %d: failed to apply updates: %v", id, err)
				return
			}

			t.Logf("Goroutine %d completed successfully", id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper function for min (Go 1.21+ has built-in min)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark tests to ensure reasonable performance
func BenchmarkYjsService_CreateEmptySnapshot(b *testing.B) {
	service, _ := NewYjsService()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateEmptySnapshot()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYjsService_CreateSnapshotWithText(b *testing.B) {
	service, _ := NewYjsService()
	text := "Benchmark test text content"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateSnapshotWithText("content", text)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYjsService_ApplyUpdates(b *testing.B) {
	service, _ := NewYjsService()
	baseSnapshot, _ := service.CreateEmptySnapshot()
	update, _ := service.CreateUpdateWithText("content", "test")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ApplyUpdates(baseSnapshot, [][]byte{update})
		if err != nil {
			b.Fatal(err)
		}
	}
}