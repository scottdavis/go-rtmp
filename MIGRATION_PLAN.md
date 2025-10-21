# RTMP Library Enhancement Migration Plan

## Overview

This document outlines the plan to migrate from the original RTMP library to the enhanced version that solves stream management and error handling issues.

## What We've Accomplished

I've created an enhanced fork of the RTMP library (`/home/sdavis/code/rtmp-enhanced`) that addresses the core issues you were facing with stream management and error handling.

## Key Problems Solved

### 1. **Missing Stream Context in Data Handlers**
- **Problem**: `OnVideo` and `OnAudio` had no way to know which stream the data belonged to
- **Solution**: Added stream context parameter to all data handlers
- **Impact**: Eliminates need for global state management and complex workarounds

### 2. **Hidden Panic Recovery**
- **Problem**: Panics were being silently converted to errors without proper logging
- **Solution**: Enhanced panic recovery with full stack traces and OnError callback integration
- **Impact**: Much better debugging and error visibility

### 3. **Poor Error Handling**
- **Problem**: No structured way to handle errors, making debugging difficult
- **Solution**: Added `OnError` callback to handler interface
- **Impact**: Handlers can now override error behavior and implement custom error handling

## Key Changes Made

### 1. Enhanced Handler Interface
```go
// Before
OnVideo(timestamp uint32, payload io.Reader) error
OnAudio(timestamp uint32, payload io.Reader) error

// After  
OnVideo(ctx *StreamContext, timestamp uint32, payload io.Reader) error
OnAudio(ctx *StreamContext, timestamp uint32, payload io.Reader) error
OnError(ctx *StreamContext, err error) error  // NEW
```

### 2. Enhanced Stream Context
```go
type StreamContext struct {
    StreamID       uint32
    StreamName     string
    PublishingName string
    App            string
    Metadata       map[string]interface{}  // NEW
}

// Helper methods
func (ctx *StreamContext) GetMetadata(key string) (interface{}, bool)
func (ctx *StreamContext) SetMetadata(key string, value interface{})
```

### 3. Better Panic Recovery
- Full stack traces in logs
- OnError callback integration
- Handler can override panic behavior
- Rich error context

### 4. Enhanced Error Logging
- Detailed error context (StreamID, ChunkStreamID, Timestamp, MessageType)
- Better debugging information
- Structured error handling

## Files Modified

- `handler.go` - Enhanced interface with stream context and OnError
- `default_handler.go` - Updated default implementations  
- `context.go` - Enhanced StreamContext with metadata
- `conn.go` - Improved error handling and panic recovery
- `chunk_streamer.go` - Better panic recovery
- `server_data_publish_handler.go` - Pass stream context to handlers

## Example Implementation

Created `example/enhanced_error_handling/` showing:
- Stream context usage
- OnError callback implementation  
- Stream-specific state management
- Enhanced error handling

## Benefits for Your Use Case

### 1. **Eliminates Global State Issues**
- No more `streamPaths` map workarounds
- No more shared handler state problems
- Each stream gets its own context

### 2. **Better Error Handling**
- Panics are properly logged with stack traces
- OnError callback lets you handle errors gracefully
- Rich debugging information

### 3. **Simpler Code**
- Direct access to stream information in OnVideo/OnAudio
- No complex workarounds needed
- Clean separation of concerns

## Migration Steps

### Step 1: Backup Current Code
```bash
# Backup your current project
cp -r /home/sdavis/code/oshi-rtmp-muxer /home/sdavis/code/oshi-rtmp-muxer-backup
```

### Step 2: Update Dependencies
```bash
# Replace the RTMP library dependency
cd /home/sdavis/code/oshi-rtmp-muxer
# Update go.mod to use the enhanced library
```

### Step 3: Update Handler Signatures
Replace your current handler methods:

```go
// OLD - Current implementation
func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {
    // Complex workarounds with streamPaths map
    var currentPath string
    if h.path != "" {
        currentPath = h.path
    } else {
        for _, path := range h.streamPaths {
            currentPath = path
            break
        }
    }
    // ... rest of logic
}

// NEW - Enhanced implementation
func (h *Handler) OnVideo(ctx *rtmp.StreamContext, timestamp uint32, payload io.Reader) error {
    // Direct access to stream context - no workarounds needed!
    streamID := ctx.StreamID
    
    // Process video for this specific stream
    // No global state management required
    return nil
}
```

### Step 4: Add OnError Method
```go
func (h *Handler) OnError(ctx *rtmp.StreamContext, err error) error {
    if ctx != nil {
        // Stream-specific error
        h.logger.Printf("Stream %d error: %v", ctx.StreamID, err)
        
        // Handle different error types
        switch {
        case strings.Contains(err.Error(), "timeout"):
            // Handle timeout errors
            return nil // Ignore and continue
        case strings.Contains(err.Error(), "format"):
            // Handle format errors
            return nil // Ignore and continue
        default:
            // Let other errors propagate
            return err
        }
    } else {
        // Connection-level error
        h.logger.Printf("Connection error: %v", err)
        return err
    }
}
```

### Step 5: Remove Workarounds
Remove these from your current handler:
- `streamPaths map[uint32]string` field
- Complex path tracking logic in OnVideo
- Global state management
- `sendVideoForStream` method (if no longer needed)

### Step 6: Update OnPublish
```go
func (h *Handler) OnPublish(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
    // Store stream-specific information in context
    ctx.StreamName = cmd.PublishingName
    ctx.PublishingName = cmd.PublishingName
    
    // Initialize stream metadata
    ctx.Metadata = make(map[string]interface{})
    ctx.SetMetadata("publish_time", time.Now())
    ctx.SetMetadata("frame_count", 0)
    
    // Create channels for this stream (if needed)
    // No global state required
    
    return nil
}
```

### Step 7: Test the Changes
```bash
# Run your existing tests
go test ./server/... -v

# Check for any compilation errors
go build -o oshi-rtmp-muxer

# Test with real RTMP streams
```

## Expected Benefits

### 1. **Eliminates "OnVideo called before path is set" Errors**
- Stream context is always available in OnVideo/OnAudio
- No more timing issues with path setting
- No more empty paths in logs

### 2. **Simplifies Code**
- Remove complex `streamPaths` map
- Remove `sendVideoForStream` workarounds
- Direct stream context access

### 3. **Better Error Handling**
- Panics are properly logged with stack traces
- OnError callback for graceful error handling
- Rich debugging information

### 4. **Stream Isolation**
- Each stream has its own context
- No shared state between streams
- Cleaner architecture

## Rollback Plan

If issues arise:
1. Restore from backup: `cp -r /home/sdavis/code/oshi-rtmp-muxer-backup /home/sdavis/code/oshi-rtmp-muxer`
2. Revert go.mod changes
3. Continue with original implementation

## Testing Checklist

- [ ] Handler compiles with new signatures
- [ ] OnError callback works correctly
- [ ] Stream context is available in OnVideo/OnAudio
- [ ] No more "path not set" errors
- [ ] Panic recovery works with proper logging
- [ ] Existing tests pass
- [ ] Real RTMP streams work correctly

## Next Session Actions

1. **Start with Step 1**: Backup current code
2. **Update dependencies**: Point to enhanced library
3. **Update handler signatures**: Add stream context parameters
4. **Add OnError method**: Implement error handling
5. **Remove workarounds**: Clean up global state management
6. **Test thoroughly**: Verify everything works

The enhanced library should solve your original issues while providing a much cleaner and more maintainable codebase.
