//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/scottdavis/go-rtmp/internal"
	"github.com/scottdavis/go-rtmp/message"
)

var _ stateHandler = (*serverDataPublishHandler)(nil)

// serverDataPublishHandler Handle data messages from a publisher at server side.
//
//	transitions:
//	  | _ -> self
type serverDataPublishHandler struct {
	sh *streamHandler
}

func (h *serverDataPublishHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		ctx := h.createStreamContext()
		return h.sh.stream.userHandler().OnAudio(ctx, timestamp, msg.Payload)

	case *message.VideoMessage:
		ctx := h.createStreamContext()
		return h.sh.stream.userHandler().OnVideo(ctx, timestamp, msg.Payload)

	default:
		return internal.ErrPassThroughMsg
	}
}

// createStreamContext creates a StreamContext with connection information
func (h *serverDataPublishHandler) createStreamContext() *StreamContext {
	ctx := &StreamContext{
		StreamID: h.sh.stream.streamID,
	}

	// Try to get connection information if available
	if h.sh.stream.conn != nil && h.sh.stream.conn.rwc != nil {
		// Try different ways to get the remote address
		if netConn, ok := h.sh.stream.conn.rwc.(interface{ RemoteAddr() string }); ok {
			ctx.RemoteAddr = netConn.RemoteAddr()
		} else if netConn, ok := h.sh.stream.conn.rwc.(interface{ RemoteAddr() net.Addr }); ok {
			ctx.RemoteAddr = netConn.RemoteAddr().String()
		}

		// Create a unique connection ID using UUID for guaranteed uniqueness
		// Format: "UUID-STREAMID" (e.g., "550e8400-e29b-41d4-a716-446655440000-1")
		connectionUUID := uuid.New().String()
		ctx.ConnectionID = fmt.Sprintf("%s-%d", connectionUUID, h.sh.stream.streamID)
	}

	return ctx
}

func (h *serverDataPublishHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	switch data := body.(type) {
	case *message.NetStreamSetDataFrame:
		ctx := h.createStreamContext()
		return h.sh.stream.userHandler().OnSetDataFrame(ctx, timestamp, data)

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverDataPublishHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}
