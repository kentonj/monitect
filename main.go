package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// publishToStream publishes to the stream,
// if the channel is full, it reads the last message, then republishes
func publishToStream(ch chan []byte, data []byte) {
	select {
	case ch <- data:
	default:
		<-ch
		ch <- data
	}
}

type StreamClient struct {
	streamSize    int
	baseStream    chan []byte
	clientStreams map[string]chan []byte
	mu            sync.RWMutex
}

func NewStreamClient(streamSize int) *StreamClient {
	baseStream := make(chan []byte, streamSize)
	clientStreams := make(map[string]chan []byte)
	return &StreamClient{
		streamSize:    streamSize,
		baseStream:    baseStream,
		clientStreams: clientStreams,
	}
}

// Publish sends data to the base stream, and then fans out to all client streams
func (s *StreamClient) Publish(data []byte) {
	publishToStream(s.baseStream, data)
	s.sendToClientStreams(data)
}

// sendToClientStreams sends the data to each of the client streams
func (s *StreamClient) sendToClientStreams(data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, clientStream := range s.clientStreams {
		publishToStream(clientStream, data)
	}
}

// AddClient adds a client
func (s *StreamClient) AddClient(clientId string) {
	clientStream := make(chan []byte, s.streamSize)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientStreams[clientId] = clientStream
}

// RemoveClient removes a client
func (s *StreamClient) RemoveClient(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("removing client %s", clientId)
	close(s.clientStreams[clientId])
	delete(s.clientStreams, clientId)
}

// ClientStream returns the stream for the specified client
func (s *StreamClient) ClientStream(clientId string) (chan []byte, bool) {
	stream, found := s.clientStreams[clientId]
	return stream, found
}

// Consume reads from a client stream for a while, then closes the stream
func (s *StreamClient) Consume(clientId string) {
	clientStream, found := s.ClientStream(clientId)
	if !found {
		return
	}
	for i := 0; i < 5; i++ {
		d := <-clientStream
		time.Sleep(15 * time.Millisecond)
		log.Printf("%d: consumed the data %s", i, string(d))
	}
	s.RemoveClient(clientId)
}

// readStream
func (s *Server) readStream(w http.ResponseWriter, r *http.Request) {
	clientId := "testing"
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`unable to upgrade to websocket connection`))
		return
	}
	s.StreamClient.AddClient(clientId)
	defer s.StreamClient.RemoveClient(clientId)
	stream, found := s.StreamClient.ClientStream(clientId)
	if !found {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`stream not found for client, this should never happen`))
		return
	}
	for msg := range stream {
		err := ws.WriteMessage(1, msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`unable to write message to websocket`))
			return
		}
		log.Printf("published message to websocket %s", string(msg))
	}
}

type Server struct {
	websocket.Upgrader
	router       *mux.Router
	StreamClient *StreamClient
}

func (s *Server) addRoutes() {
	s.router.HandleFunc("/consume/{clientId}", s.readStream)
}

func NewServer(streamClient *StreamClient) *Server {
	server := Server{Upgrader: websocket.Upgrader{}, router: mux.NewRouter(), StreamClient: streamClient}
	return &server
}

func (s *Server) Start() {
	log.Print("starting server")
	s.addRoutes()
	http.ListenAndServe(":8080", nil)
}

func main() {
	streamer := NewStreamClient(1000)
	go func() {
		log.Print("starting to publish to streamer in the background")
		for i := 0; i < 10000; i++ {
			streamer.Publish([]byte(fmt.Sprintf("message: %d", i)))
			time.Sleep(1 * time.Millisecond)
		}

	}()
	server := NewServer(streamer)
	server.Start()
}
