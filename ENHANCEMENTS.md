# RTMP Library Enhancements

This document describes the improvements made to the RTMP library to better support stream management and error handling.

## Key Enhancements

### 1. Stream Context in Data Handlers

**Problem**: `OnVideo` and `OnAudio` methods didn't receive stream context, making it impossible to know which stream the data belonged to.

**Solution**: Enhanced the handler interface to include stream context:

```go
// Before
OnVideo(timestamp uint32, payload io.Reader) error
OnAudio(timestamp uint32, payload io.Reader) error

// After
OnVideo(ctx *StreamContext, timestamp uint32, payload io.Reader) error
OnAudio(ctx *StreamContext, timestamp uint32, payload io.Reader) error
```

**Benefits**:
- ✅ Eliminates need for global state management
- ✅ Enables per-stream isolation
- ✅ Simplifies debugging and logging
- ✅ Makes stream-specific logic possible

### 2. Enhanced Stream Context

**Problem**: Limited stream context information available to handlers.

**Solution**: Extended `StreamContext` with more useful fields:

```go
type StreamContext struct {
    StreamID       uint32
    StreamName     string
    PublishingName string
    App            string
    // Metadata storage for custom data
    Metadata       map[string]interface{}
}

// Helper methods
func (ctx *StreamContext) GetMetadata(key string) (interface{}, bool)
func (ctx *StreamContext) SetMetadata(key string, value interface{})
```

**Benefits**:
- ✅ Rich stream metadata storage
- ✅ Easy access to stream information
- ✅ Custom data storage per stream

### 3. OnError Callback

**Problem**: Errors and panics were being hidden or poorly handled.

**Solution**: Added `OnError` callback to the handler interface:

```go
type Handler interface {
    // ... existing methods ...
    OnError(ctx *StreamContext, err error) error
}
```

**Benefits**:
- ✅ Structured error handling
- ✅ Ability to override error behavior
- ✅ Better debugging and logging
- ✅ Graceful error recovery

### 4. Improved Panic Recovery

**Problem**: Panics were being silently converted to errors without proper logging.

**Solution**: Enhanced panic recovery with better logging and error context:

```go
defer func() {
    if r := recover(); r != nil {
        // Better panic recovery with stack traces
        var errTmp error
        if panicErr, ok := r.(error); ok {
            errTmp = errors.WithStack(panicErr)
        } else {
            errTmp = errors.Errorf("Panic in message loop: %+v", r)
        }
        
        // Log with full context
        if c.logger != nil {
            c.logger.Errorf("PANIC RECOVERED: %+v", errTmp)
        }
        
        // Use OnError callback
        if c.handler != nil {
            if handlerErr := c.handler.OnError(nil, errTmp); handlerErr != nil {
                err = handlerErr
                return
            }
        }
        
        err = errTmp
    }
}()
```

**Benefits**:
- ✅ Full stack traces in logs
- ✅ Handler can override panic behavior
- ✅ Better debugging information
- ✅ Graceful error recovery

### 5. Enhanced Error Logging

**Problem**: Errors lacked context for debugging.

**Solution**: Added detailed error logging with full context:

```go
// Log all errors with more context for debugging
if c.logger != nil {
    c.logger.Errorf("Error handling message: StreamID=%d, ChunkStreamID=%d, Timestamp=%d, MessageType=%T, Error=%+v", 
        cmsg.StreamID, chunkStreamID, timestamp, cmsg.Message, err)
}
```

**Benefits**:
- ✅ Rich debugging information
- ✅ Easy to trace error sources
- ✅ Better error analysis

## Usage Examples

### Basic Enhanced Handler

```go
type EnhancedHandler struct {
    rtmp.DefaultHandler
    logger *log.Logger
}

func (h *EnhancedHandler) OnVideo(ctx *rtmp.StreamContext, timestamp uint32, payload io.Reader) error {
    // Now we have direct access to stream context!
    h.logger.Printf("Video frame: StreamID=%d, Timestamp=%d", ctx.StreamID, timestamp)
    
    // Store stream-specific metadata
    ctx.SetMetadata("last_video_timestamp", timestamp)
    
    // Process video data...
    return nil
}

func (h *EnhancedHandler) OnError(ctx *rtmp.StreamContext, err error) error {
    if ctx != nil {
        h.logger.Printf("Stream error: StreamID=%d, Error=%v", ctx.StreamID, err)
        
        // Handle different error types
        switch {
        case err.Error() == "connection timeout":
            // Retry logic
            return nil // Ignore error
        default:
            return err // Let error propagate
        }
    } else {
        // Connection-level error
        h.logger.Printf("Connection error: %v", err)
        return err
    }
}
```

### Stream-Specific State Management

```go
func (h *EnhancedHandler) OnPublish(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
    // Initialize stream-specific state
    ctx.StreamName = cmd.PublishingName
    ctx.Metadata = make(map[string]interface{})
    ctx.SetMetadata("publish_time", time.Now())
    ctx.SetMetadata("frame_count", 0)
    
    return nil
}
```

## Migration Guide

### For Existing Handlers

1. **Update method signatures**:
   ```go
   // Old
   func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error
   
   // New
   func (h *Handler) OnVideo(ctx *StreamContext, timestamp uint32, payload io.Reader) error
   ```

2. **Add OnError method**:
   ```go
   func (h *Handler) OnError(ctx *StreamContext, err error) error {
       // Handle errors as needed
       return err // Default: let error propagate
   }
   ```

3. **Use stream context**:
   ```go
   func (h *Handler) OnVideo(ctx *StreamContext, timestamp uint32, payload io.Reader) error {
       // Access stream information
       streamID := ctx.StreamID
       // Use stream-specific logic
   }
   ```

## Benefits Summary

1. **Eliminates Global State**: No more complex workarounds for stream isolation
2. **Better Error Handling**: Structured error handling with OnError callback
3. **Enhanced Debugging**: Rich logging and stack traces
4. **Stream Isolation**: Each stream can have its own state and logic
5. **Backward Compatibility**: Default implementations maintain compatibility
6. **Flexible Error Recovery**: Handlers can override error behavior

## Files Modified

- `handler.go` - Enhanced interface with stream context and OnError
- `default_handler.go` - Updated default implementations
- `context.go` - Enhanced StreamContext with metadata
- `conn.go` - Improved error handling and panic recovery
- `chunk_streamer.go` - Better panic recovery
- `server_data_publish_handler.go` - Pass stream context to handlers

## Example Implementation

See `example/enhanced_error_handling/` for a complete example demonstrating:
- Stream context usage
- OnError callback implementation
- Stream-specific state management
- Enhanced error handling
