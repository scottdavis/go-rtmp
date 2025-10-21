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
	"github.com/pkg/errors"

	"github.com/scottdavis/go-rtmp/internal"
	"github.com/scottdavis/go-rtmp/message"
)

var _ stateHandler = (*serverControlConnectedHandler)(nil)

// serverControlConnectedHandler Handle control messages from a client at server side.
//
//	transitions:
//	  | "createStream" -> spawn! serverDataInactiveHandler
//	  | _              -> self
type serverControlConnectedHandler struct {
	sh *streamHandler
}

func (h *serverControlConnectedHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) (err error) {
	l := h.sh.Logger()
	tID := cmdMsg.TransactionID

	switch cmd := body.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)
		defer func() {
			if err != nil {
				result := h.newCreateStreamErrorResult()

				l.Infof("CreateStream(Error): ResponseBody = %#v, Err = %+v", result, err)
				if err1 := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err1 != nil {
					err = errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
				}
			}
		}()

		// Create a basic stream context for connection-level operations
		ctx := h.createStreamContext()
		if err := h.sh.stream.userHandler().OnCreateStream(ctx, timestamp, cmd); err != nil {
			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		newStream, err := h.sh.stream.streams().conn.streams.CreateIfAvailable()
		if err != nil {
			l.Errorf("Failed to create stream: Err = %+v", err)

			result := h.newCreateStreamErrorResult()
			if err1 := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err1 != nil {
				return errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
			}

			return nil // Keep the connection
		}
		newStream.handler.ChangeState(streamStateServerInactive)

		result := h.newCreateStreamSuccessResult(newStream.streamID)
		if err := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err != nil {
			_ = h.sh.stream.streams().Delete(newStream.streamID) // TODO: error handling
			return err
		}

		l.Infof("Stream created...: NewStreamID = %d", newStream.streamID)

		return nil

	case *message.NetStreamDeleteStream:
		l.Infof("Stream deleting...: TargetStreamID = %d", cmd.StreamID)

		ctx := &StreamContext{StreamID: cmd.StreamID}
		if err := h.sh.stream.userHandler().OnDeleteStream(ctx, timestamp, cmd); err != nil {
			return err
		}

		if err := h.sh.stream.streams().Delete(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		l.Infof("Stream deleted: TargetStreamID = %d", cmd.StreamID)

		return nil

	case *message.NetConnectionReleaseStream:
		l.Infof("Release stream...: StreamName = %s", cmd.StreamName)

		ctx := h.createStreamContext()
		if err := h.sh.stream.userHandler().OnReleaseStream(ctx, timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCPublish:
		l.Infof("FCPublish stream...: StreamName = %s", cmd.StreamName)

		ctx := h.createStreamContext()
		if err := h.sh.stream.userHandler().OnFCPublish(ctx, timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCUnpublish:
		l.Infof("FCUnpublish stream...: StreamName = %s", cmd.StreamName)

		ctx := h.createStreamContext()
		if err := h.sh.stream.userHandler().OnFCUnpublish(ctx, timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverControlConnectedHandler) newCreateStreamSuccessResult(
	streamID uint32,
) *message.NetConnectionCreateStreamResult {
	return &message.NetConnectionCreateStreamResult{
		StreamID: streamID,
	}
}

func (h *serverControlConnectedHandler) newCreateStreamErrorResult() *message.NetConnectionCreateStreamResult {
	return nil
}

// createStreamContext creates a StreamContext with connection information
func (h *serverControlConnectedHandler) createStreamContext() *StreamContext {
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
