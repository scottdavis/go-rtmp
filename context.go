//
// Copyright (c) 2021- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

type StreamContext struct {
	StreamID       uint32
	StreamName     string
	PublishingName string
	App            string
	// Connection-level context for unique identification
	ConnectionID string // Unique identifier for the connection
	RemoteAddr   string // Remote address of the connection
	// Metadata storage for custom data
	Metadata map[string]interface{}
}

// GetMetadata retrieves a metadata value by key
func (ctx *StreamContext) GetMetadata(key string) (interface{}, bool) {
	if ctx.Metadata == nil {
		return nil, false
	}
	value, exists := ctx.Metadata[key]
	return value, exists
}

// SetMetadata sets a metadata value by key
func (ctx *StreamContext) SetMetadata(key string, value interface{}) {
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]interface{})
	}
	ctx.Metadata[key] = value
}
