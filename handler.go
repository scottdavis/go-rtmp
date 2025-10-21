//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"io"

	"github.com/yutopp/go-rtmp/message"
)

type Handler interface {
	OnServe(conn *Conn)
	OnConnect(ctx *StreamContext, timestamp uint32, cmd *message.NetConnectionConnect) error
	OnCreateStream(ctx *StreamContext, timestamp uint32, cmd *message.NetConnectionCreateStream) error
	OnReleaseStream(ctx *StreamContext, timestamp uint32, cmd *message.NetConnectionReleaseStream) error
	OnDeleteStream(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamDeleteStream) error
	OnPublish(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error
	OnPlay(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamPlay) error
	OnFCPublish(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamFCPublish) error
	OnFCUnpublish(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamFCUnpublish) error
	OnSetDataFrame(ctx *StreamContext, timestamp uint32, data *message.NetStreamSetDataFrame) error
	OnAudio(ctx *StreamContext, timestamp uint32, payload io.Reader) error
	OnVideo(ctx *StreamContext, timestamp uint32, payload io.Reader) error
	OnUnknownMessage(ctx *StreamContext, timestamp uint32, msg message.Message) error
	OnUnknownCommandMessage(ctx *StreamContext, timestamp uint32, cmd *message.CommandMessage) error
	OnUnknownDataMessage(ctx *StreamContext, timestamp uint32, data *message.DataMessage) error
	// ENHANCED: Add error callback for better error handling
	OnError(ctx *StreamContext, err error) error
	OnClose()
}
