package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"
)

// EnhancedHandler demonstrates the new OnError callback and stream context features
type EnhancedHandler struct {
	rtmp.DefaultHandler
	logger *log.Logger
}

func NewEnhancedHandler() *EnhancedHandler {
	return &EnhancedHandler{
		logger: log.New(log.Writer(), "[Enhanced] ", log.LstdFlags),
	}
}

// OnError demonstrates the new error handling callback
func (h *EnhancedHandler) OnError(ctx *rtmp.StreamContext, err error) error {
	if ctx != nil {
		h.logger.Printf("Stream error: StreamID=%d, Error=%v", ctx.StreamID, err)

		// You can decide how to handle different types of errors
		switch {
		case err.Error() == "connection timeout":
			// For timeout errors, we might want to retry
			h.logger.Printf("Timeout detected for stream %d, attempting recovery", ctx.StreamID)
			return nil // Return nil to ignore the error and continue

		case err.Error() == "invalid video format":
			// For format errors, we might want to log but continue
			h.logger.Printf("Invalid video format for stream %d, continuing anyway", ctx.StreamID)
			return nil // Ignore the error

		default:
			// For other errors, let them propagate
			h.logger.Printf("Unhandled error for stream %d: %v", ctx.StreamID, err)
			return err
		}
	} else {
		// Connection-level error (no stream context)
		h.logger.Printf("Connection error: %v", err)

		// For connection-level errors, you might want to:
		// - Log the error
		// - Send a notification
		// - Attempt reconnection
		// - Or just let it propagate
		return err
	}
}

// OnVideo demonstrates the new stream context feature
func (h *EnhancedHandler) OnVideo(ctx *rtmp.StreamContext, timestamp uint32, payload io.Reader) error {
	// Now we have direct access to the stream context!
	h.logger.Printf("Video frame: StreamID=%d, Timestamp=%d", ctx.StreamID, timestamp)

	// You can store stream-specific metadata
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]interface{})
	}
	ctx.SetMetadata("last_video_timestamp", timestamp)
	if frameCount, exists := ctx.GetMetadata("frame_count"); exists {
		ctx.SetMetadata("frame_count", frameCount.(int)+1)
	} else {
		ctx.SetMetadata("frame_count", 1)
	}

	// Process the video data
	data, err := io.ReadAll(payload)
	if err != nil {
		return fmt.Errorf("failed to read video data: %w", err)
	}

	h.logger.Printf("Processed video frame: StreamID=%d, Size=%d bytes", ctx.StreamID, len(data))

	return nil
}

// OnAudio demonstrates the new stream context feature
func (h *EnhancedHandler) OnAudio(ctx *rtmp.StreamContext, timestamp uint32, payload io.Reader) error {
	// Now we have direct access to the stream context!
	h.logger.Printf("Audio frame: StreamID=%d, Timestamp=%d", ctx.StreamID, timestamp)

	// Process the audio data
	data, err := io.ReadAll(payload)
	if err != nil {
		return fmt.Errorf("failed to read audio data: %w", err)
	}

	h.logger.Printf("Processed audio frame: StreamID=%d, Size=%d bytes", ctx.StreamID, len(data))

	return nil
}

// OnPublish demonstrates enhanced stream context
func (h *EnhancedHandler) OnPublish(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	h.logger.Printf("Stream published: StreamID=%d, Name=%s", ctx.StreamID, cmd.PublishingName)

	// Store stream-specific information
	ctx.StreamName = cmd.PublishingName
	ctx.PublishingName = cmd.PublishingName

	// Initialize stream metadata
	ctx.Metadata = make(map[string]interface{})
	ctx.SetMetadata("publish_time", time.Now())
	ctx.SetMetadata("frame_count", 0)

	return nil
}

// OnDeleteStream demonstrates cleanup with stream context
func (h *EnhancedHandler) OnDeleteStream(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamDeleteStream) error {
	h.logger.Printf("Stream deleted: StreamID=%d", cmd.StreamID)

	// You can access stream metadata for cleanup
	if ctx != nil {
		h.logger.Printf("Stream context available: StreamID=%d", ctx.StreamID)
		// Access metadata for cleanup
		if publishTime, exists := ctx.GetMetadata("publish_time"); exists {
			h.logger.Printf("Stream was published at: %v", publishTime)
		}
	}

	return nil
}

// OnConnect demonstrates connection-level error handling
func (h *EnhancedHandler) OnConnect(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetConnectionConnect) error {
	h.logger.Printf("Client connected")
	return nil
}

// OnClose demonstrates cleanup
func (h *EnhancedHandler) OnClose() {
	h.logger.Printf("Connection closed")
}
