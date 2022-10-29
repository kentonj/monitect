package stream

import (
	"fmt"
	"log"
	"sync"
)

var (
	CoreStreamName = "_CORE"
)

// Manager keeps a core stream and a collection of ClientStream
type Manager struct {
	streamSize    int
	coreStream    *Stream
	clientStreams map[string]*Stream
	mu            sync.RWMutex
}

func NewManager(streamSize int) *Manager {
	m := Manager{
		streamSize:    streamSize,
		coreStream:    NewStream(CoreStreamName, streamSize),
		clientStreams: make(map[string]*Stream),
	}
	// start fanning out
	go m.fanout()
	return &m
}

// Fanout reads from the core stream, and distributes to all clients, until the core stream is closed
func (m *Manager) fanout() {
	for d := range m.coreStream.ch {
		func() {
			m.mu.RLock()
			defer m.mu.RUnlock()
			for _, clientStream := range m.clientStreams {
				clientStream.Send(d)
			}
		}()
	}
}

// Send adds a message to the coreStream
func (m *Manager) Send(d []byte) {
	m.coreStream.Send(d)
}

// AddClient adds a client
func (m *Manager) AddClient(clientId string) error {
	if clientId == CoreStreamName {
		return fmt.Errorf("cannot use core stream %s", CoreStreamName)
	}
	clientStream := NewStream(clientId, m.streamSize)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clientStreams[clientId] = clientStream
	log.Printf("added client %s", clientId)
	return nil
}

// RemoveClient removes a client by closing the ClientStream.ch and removing it from the map of clientStreams
func (m *Manager) RemoveClient(clientId string) error {
	if clientId == CoreStreamName {
		return fmt.Errorf("cannot remove core stream %s", CoreStreamName)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if clientStream, found := m.clientStreams[clientId]; found {
		close(clientStream.ch)
	}
	delete(m.clientStreams, clientId)
	return nil
}

// ClientStream returns the ClientStream's channel for reading,
func (m *Manager) ClientStream(clientId string) (*Stream, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clientStream, found := m.clientStreams[clientId]
	return clientStream, found
}
