package stream

import (
	"fmt"
	"log"
	"sync"
)

// publishToStream publishes to the stream,
// if the channel is full, it reads the last message (opening up a spot on the channel), then republishes
func publishToStream(ch chan []byte, data []byte) {
	select {
	case ch <- data:
	default:
		<-ch
		ch <- data
	}
}

type Stream struct {
	size          int
	baseStream    chan []byte
	clientStreams map[string]chan []byte
	mu            sync.RWMutex
}

func NewStream(streamSize int) *Stream {
	baseStream := make(chan []byte, streamSize)
	clientStreams := make(map[string]chan []byte)
	return &Stream{
		size:          streamSize,
		baseStream:    baseStream,
		clientStreams: clientStreams,
	}
}

// Publish sends data to the base stream, and then fans out to all client streams
func (s *Stream) Publish(data []byte) {
	publishToStream(s.baseStream, data)
	s.sendToClientStreams(data)
}

// sendToClientStreams sends the data to each of the client streams
func (s *Stream) sendToClientStreams(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, clientStream := range s.clientStreams {
		publishToStream(clientStream, data)
	}
}

// AddClient adds a client to the specified stream
func (s *Stream) AddClient(clientId string) error {
	if _, found := s.clientStreams[clientId]; found {
		return fmt.Errorf("client %s already exists", clientId)
	}
	clientStream := make(chan []byte, s.size)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientStreams[clientId] = clientStream
	log.Printf("added client %s (total clients %d)", clientId, len(s.clientStreams))
	return nil
}

// RemoveClient removes a client
func (s *Stream) RemoveClient(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.clientStreams[clientId])
	delete(s.clientStreams, clientId)
	log.Printf("removed client %s (total clients %d)", clientId, len(s.clientStreams))
}

// Client returns the stream for the specified client
func (s *Stream) Client(clientId string) (chan []byte, bool) {
	stream, found := s.clientStreams[clientId]
	return stream, found
}
