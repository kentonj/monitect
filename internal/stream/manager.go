package stream

import "fmt"

// Manager is a named map of Stream pointers
type Manager struct {
	streamSize int
	streams    map[string]*Stream
}

// NewManager creates and returns a pointer to a StreamManager
func NewManager(streamSize int) *Manager {
	return &Manager{streamSize: streamSize, streams: make(map[string]*Stream)}
}

// Add creates and allocates named streams, so that they are ready for data to be published to them
func (m *Manager) Add(streamId string) error {
	if _, found := m.streams[streamId]; found {
		return fmt.Errorf("stream with id %s already exists", streamId)
	}
	stream := NewStream(m.streamSize)
	m.streams[streamId] = stream
	return nil
}

// Remove removes the stream from the list of streams
func (m *Manager) Remove(streamId string) {
	delete(m.streams, streamId)
}

// Stream returns the Stream associated with the specified streamId
func (m *Manager) Stream(streamId string) (*Stream, bool) {
	stream, found := m.streams[streamId]
	return stream, found
}
