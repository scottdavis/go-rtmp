package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"
)

// TestEnhancedHandler tests the new OnError callback and stream context features
func TestEnhancedHandler(t *testing.T) {
	handler := NewEnhancedHandler()
	
	// Test OnError callback
	t.Run("OnError callback", func(t *testing.T) {
		// Test with stream context
		ctx := &rtmp.StreamContext{StreamID: 123}
		err := fmt.Errorf("test error")
		
		result := handler.OnError(ctx, err)
		if result != err {
			t.Errorf("Expected error to propagate, got: %v", result)
		}
		
		// Test without stream context (connection-level error)
		result = handler.OnError(nil, err)
		if result != err {
			t.Errorf("Expected error to propagate, got: %v", result)
		}
	})
	
	// Test OnVideo with stream context
	t.Run("OnVideo with stream context", func(t *testing.T) {
		ctx := &rtmp.StreamContext{StreamID: 456}
		payload := bytes.NewReader([]byte("test video data"))
		
		err := handler.OnVideo(ctx, 1000, payload)
		if err != nil {
			t.Errorf("OnVideo failed: %v", err)
		}
		
		// Check that metadata was set
		if ctx.Metadata == nil {
			t.Error("Expected metadata to be initialized")
		}
		
		if _, exists := ctx.GetMetadata("last_video_timestamp"); !exists {
			t.Error("Expected last_video_timestamp to be set")
		}
	})
	
	// Test OnAudio with stream context
	t.Run("OnAudio with stream context", func(t *testing.T) {
		ctx := &rtmp.StreamContext{StreamID: 789}
		payload := bytes.NewReader([]byte("test audio data"))
		
		err := handler.OnAudio(ctx, 2000, payload)
		if err != nil {
			t.Errorf("OnAudio failed: %v", err)
		}
	})
	
	// Test OnPublish with stream context
	t.Run("OnPublish with stream context", func(t *testing.T) {
		ctx := &rtmp.StreamContext{StreamID: 999}
		cmd := &message.NetStreamPublish{
			PublishingName: "test-stream",
		}
		
		err := handler.OnPublish(ctx, 3000, cmd)
		if err != nil {
			t.Errorf("OnPublish failed: %v", err)
		}
		
		// Check that stream information was set
		if ctx.StreamName != "test-stream" {
			t.Errorf("Expected StreamName to be 'test-stream', got: %s", ctx.StreamName)
		}
		
		if ctx.PublishingName != "test-stream" {
			t.Errorf("Expected PublishingName to be 'test-stream', got: %s", ctx.PublishingName)
		}
		
		// Check that metadata was initialized
		if ctx.Metadata == nil {
			t.Error("Expected metadata to be initialized")
		}
		
		if _, exists := ctx.GetMetadata("publish_time"); !exists {
			t.Error("Expected publish_time to be set")
		}
		
		if frameCount, exists := ctx.GetMetadata("frame_count"); !exists || frameCount != 0 {
			t.Errorf("Expected frame_count to be 0, got: %v", frameCount)
		}
	})
}

// TestStreamContextMetadata tests the metadata functionality
func TestStreamContextMetadata(t *testing.T) {
	ctx := &rtmp.StreamContext{StreamID: 123}
	
	// Test setting and getting metadata
	ctx.SetMetadata("test_key", "test_value")
	
	value, exists := ctx.GetMetadata("test_key")
	if !exists {
		t.Error("Expected metadata to exist")
	}
	
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got: %v", value)
	}
	
	// Test getting non-existent metadata
	_, exists = ctx.GetMetadata("non_existent")
	if exists {
		t.Error("Expected metadata to not exist")
	}
	
	// Test setting multiple values
	ctx.SetMetadata("count", 42)
	ctx.SetMetadata("active", true)
	
	if count, _ := ctx.GetMetadata("count"); count != 42 {
		t.Errorf("Expected count to be 42, got: %v", count)
	}
	
	if active, _ := ctx.GetMetadata("active"); active != true {
		t.Errorf("Expected active to be true, got: %v", active)
	}
}

// TestErrorHandling tests the enhanced error handling
func TestErrorHandling(t *testing.T) {
	handler := NewEnhancedHandler()
	
	// Test different error types
	testCases := []struct {
		name        string
		err         error
		expectError bool
	}{
		{
			name:        "connection timeout",
			err:         fmt.Errorf("connection timeout"),
			expectError: false, // Should be ignored
		},
		{
			name:        "invalid video format",
			err:         fmt.Errorf("invalid video format"),
			expectError: false, // Should be ignored
		},
		{
			name:        "unknown error",
			err:         fmt.Errorf("unknown error"),
			expectError: true, // Should propagate
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &rtmp.StreamContext{StreamID: 123}
			result := handler.OnError(ctx, tc.err)
			
			if tc.expectError && result == nil {
				t.Error("Expected error to propagate")
			}
			
			if !tc.expectError && result != nil {
				t.Errorf("Expected error to be ignored, got: %v", result)
			}
		})
	}
}
