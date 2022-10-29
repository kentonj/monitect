package stream

import (
	"log"
)

// Stream is meant to be assigned to a specific client, and contains an internal channel.
type Stream struct {
	clientId string
	ch       chan []byte
}

// NewStream creates a Stream with a channel of specified size
func NewStream(clientId string, size int) *Stream {
	return &Stream{clientId: clientId, ch: make(chan []byte, size)}
}

// Send adds a message to the internal channel, if it's full, then drop the first message and add the new one
func (s *Stream) Send(d []byte) {
	select {
	case s.ch <- d:
	default:
		log.Printf("client %s's channel is full, dropping oldest message", s.clientId)
		<-s.ch
		s.ch <- d
	}
}

// Poll returns as many messages as are available on the internal channel, up to the limit provided
// This may return an empty slice if no messages are available
func (s *Stream) Poll(limit int) [][]byte {
	msgs := make([][]byte, 0)
	for {
		if len(msgs) == limit {
			return msgs
		}
		select {
		case msg := <-s.ch:
			msgs = append(msgs, msg)
		default:
			// no more available messages
			return msgs
		}
	}
}

// C is a receive only channel that can be used to read from the internal channel directly
func (s *Stream) C() <-chan []byte {
	return s.ch
}
