package main

import (
	"fmt"
	"testing"
	"time"
)

func TestNewStreamClient(t *testing.T) {
	streamClient := NewStreamClient(10)
	clients := make([]string, 0)
	// can easily handle up to 100k clients with the streams replicating (10ms)
	// about 50 microseconds per 1000 clients
	for clientI := 0; clientI < 3; clientI++ {
		clients = append(clients, fmt.Sprintf("client%d", clientI))
	}
	t.Logf("got %d clients", len(clients))
	for _, c := range clients {
		streamClient.AddClient(c)
	}
	go streamClient.Consume(clients[0])
	for i := 0; i < 100; i++ {
		start := time.Now()
		streamClient.Publish([]byte(fmt.Sprintf("message: %d", i)))
		dur := time.Since(start)
		t.Logf("publish took %s", dur)
		time.Sleep(3 * time.Millisecond)
	}
}
