//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
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
		ctx := &StreamContext{StreamID: h.sh.stream.streamID}
		return h.sh.stream.userHandler().OnAudio(ctx, timestamp, msg.Payload)

	case *message.VideoMessage:
		ctx := &StreamContext{StreamID: h.sh.stream.streamID}
		return h.sh.stream.userHandler().OnVideo(ctx, timestamp, msg.Payload)

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverDataPublishHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	switch data := body.(type) {
	case *message.NetStreamSetDataFrame:
		ctx := &StreamContext{StreamID: h.sh.stream.streamID}
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
